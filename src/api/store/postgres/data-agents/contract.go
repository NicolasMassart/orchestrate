package dataagents

import (
	"context"
	"strings"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const contractDAComponent = "data-agents.contract"

type PGContract struct {
	db     pg.DB
	logger *log.Logger
}

type contractQuery struct {
	Name             string `json:"name,omitempty"`
	Tag              string `json:"tag,omitempty"`
	ABI              string `json:"abi,omitempty"`
	Bytecode         string `json:"bytecode,omitempty"`
	DeployedBytecode string `json:"deployed_bytecode,omitempty"`
}

func NewPGContract(db pg.DB) store.ContractAgent {
	return &PGContract{
		db:     db,
		logger: log.NewLogger().SetComponent(contractDAComponent),
	}
}

func (agent *PGContract) Register(ctx context.Context, contract *entities.Contract) error {
	repositoryID, err := agent.selectOrInsertRepository(ctx, &models.RepositoryModel{
		Name: contract.Name,
	})
	if err != nil {
		return err
	}

	artifactID, err := agent.selectOrInsertArtifact(ctx, &models.ArtifactModel{
		ABI:              contract.RawABI,
		Bytecode:         contract.Bytecode.String(),
		DeployedBytecode: contract.DeployedBytecode.String(),
		Codehash:         contract.CodeHash.String(),
	})
	if err != nil {
		return err
	}

	err = agent.insertTag(ctx, &models.TagModel{
		Name:         contract.Tag,
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (agent *PGContract) RegisterDeployment(ctx context.Context, chainID string, address ethcommon.Address, codeHash hexutil.Bytes) error {
	codehash := &models.CodehashModel{
		ChainID:  chainID,
		Address:  address.Hex(),
		Codehash: codeHash.String(),
	}

	// If uniqueness constraint is broken then it updates the former value
	_, err := agent.db.ModelContext(ctx, codehash).
		OnConflict("ON CONSTRAINT codehashes_chain_id_address_key DO UPDATE").
		Set("chain_id = ?chain_id").
		Set("address = ?address").
		Set("codehash = ?codehash").
		Returning("*").
		Insert()

	if err != nil {
		errMessage := "could not register contract deployment"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	return nil
}

func (agent *PGContract) FindOneByCodeHash(ctx context.Context, codeHash string) (*entities.Contract, error) {
	qContract := &contractQuery{}
	query := `
SELECT a.abi, a.bytecode, a.deployed_bytecode, r.name as name, t.name as tag 
FROM artifacts a
INNER JOIN tags t ON (a.id = t.artifact_id)
INNER JOIN repositories r ON (r.id = t.repository_id) 
WHERE a.codehash = ?
ORDER BY t.id DESC
LIMIT 1
`
	_, err := agent.db.QueryOneContext(ctx, qContract, query, codeHash)
	if err != nil {
		return nil, errors.FromError(pg.ParsePGError(err)).ExtendComponent(contractDAComponent)
	}

	return parseContract(qContract)
}

func (agent *PGContract) FindOneByAddress(ctx context.Context, address string) (*entities.Contract, error) {
	qContract := &contractQuery{}
	query := `
SELECT a.abi, a.bytecode, a.deployed_bytecode, r.name as name, t.name as tag 
FROM artifacts a
INNER JOIN tags t ON (a.id = t.artifact_id)
INNER JOIN repositories r ON (r.id = t.repository_id) 
INNER JOIN codehashes ch ON (ch.codehash = a.codehash) 
WHERE ch.address = ?
LIMIT 1
`
	_, err := agent.db.QueryOneContext(ctx, qContract, query, address)
	if err != nil {
		return nil, errors.FromError(pg.ParsePGError(err)).ExtendComponent(contractDAComponent)
	}

	return parseContract(qContract)
}

func (agent *PGContract) FindOneByNameAndTag(ctx context.Context, name, tag string) (*entities.Contract, error) {
	artifact := &models.ArtifactModel{}
	query := agent.db.ModelContext(ctx, artifact).
		Column("artifact_model.id", "abi", "bytecode", "deployed_bytecode").
		Join("JOIN tags AS t ON t.artifact_id = artifact_model.id").
		Join("JOIN repositories AS registry ON registry.id = t.repository_id").
		Where("LOWER(t.name) = LOWER(?)", tag).
		Where("LOWER(registry.name) = LOWER(?)", name)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	contract := &entities.Contract{
		Name:             name,
		Tag:              tag,
		RawABI:           artifact.ABI,
		Bytecode:         hexutil.MustDecode(artifact.Bytecode),
		DeployedBytecode: hexutil.MustDecode(artifact.DeployedBytecode),
	}

	return contract, nil
}

func (agent *PGContract) ListNames(ctx context.Context) ([]string, error) {
	var names []string
	query := agent.db.ModelContext(ctx, (*models.RepositoryModel)(nil)).
		Column("name").
		OrderExpr("lower(name)")

	err := pg.SelectColumn(ctx, query, &names)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to fetch repository names")
		}
		return nil, errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	return names, nil
}

func (agent *PGContract) ListTags(ctx context.Context, name string) ([]string, error) {
	var tags []string
	query := agent.db.ModelContext(ctx, (*models.TagModel)(nil)).
		Column("tag_model.name").
		Join("JOIN repositories AS registry ON registry.id = tag_model.repository_id").
		Where("lower(registry.name) = lower(?)", name).
		OrderExpr("lower(tag_model.name)")

	err := pg.SelectColumn(ctx, query, &tags)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to fetch tag names")
		}
		return nil, errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	return tags, nil
}

func (agent *PGContract) selectOrInsertRepository(ctx context.Context, model *models.RepositoryModel) (int, error) {
	q := agent.db.ModelContext(ctx, model).Column("id").Where("name = ?name").
		OnConflict("DO NOTHING").Returning("id")

	err := pg.SelectOrInsert(ctx, q)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to select or insert repository")
		return 0, errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	return model.ID, nil
}

func (agent *PGContract) selectOrInsertArtifact(ctx context.Context, model *models.ArtifactModel) (int, error) {
	q := agent.db.ModelContext(ctx, model).Column("id").
		Where("abi = ?abi").Where("codehash = ?codehash").
		OnConflict("DO NOTHING").Returning("id")

	err := pg.SelectOrInsert(ctx, q)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to select or insert artifact")
		return 0, errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	return model.ID, nil
}

func (agent *PGContract) insertTag(ctx context.Context, model *models.TagModel) error {
	query := agent.db.ModelContext(ctx, model).
		OnConflict("ON CONSTRAINT tags_name_repository_id_key DO UPDATE").
		Set("artifact_id = ?artifact_id")

	err := pg.InsertQuery(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert tag")
		return errors.FromError(err).ExtendComponent(contractDAComponent)
	}

	return nil
}

func parseContract(qContract *contractQuery) (*entities.Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(qContract.ABI))
	if err != nil {
		return nil, err
	}

	return &entities.Contract{
		Name:             qContract.Name,
		Tag:              qContract.Tag,
		RawABI:           qContract.ABI,
		ABI:              parsedABI,
		Bytecode:         hexutil.MustDecode(qContract.Bytecode),
		DeployedBytecode: hexutil.MustDecode(qContract.DeployedBytecode),
	}, nil
}
