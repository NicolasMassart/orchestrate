package postgres

import (
	"context"
	"strings"

	"github.com/consensys/orchestrate/src/infra/postgres"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type PGContract struct {
	client postgres.Client
	logger *log.Logger
}

type contractQuery struct {
	Name             string `json:"name,omitempty"`
	Tag              string `json:"tag,omitempty"`
	ABI              string `json:"abi,omitempty"`
	Bytecode         string `json:"bytecode,omitempty"`
	DeployedBytecode string `json:"deployed_bytecode,omitempty"`
}

var _ store.ContractAgent = &PGContract{}

func NewPGContract(client postgres.Client) *PGContract {
	return &PGContract{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.contract"),
	}
}

func (agent *PGContract) Register(ctx context.Context, contract *entities.Contract) error {
	return agent.client.RunInTransaction(ctx, func(c postgres.Client) error {
		// Insert repository
		repository := models.NewRepository(contract.Name)
		err := c.ModelContext(ctx, repository).
			Column("id").
			Where("name = ?name").
			OnConflict("DO NOTHING").
			Returning("id").
			SelectOrInsert()
		if err != nil {
			errMessage := "failed to select or insert repository"
			agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
			return errors.FromError(err).SetMessage(errMessage)
		}

		// Insert artifact
		artifact := models.NewArtifact(contract)
		err = c.ModelContext(ctx, artifact).
			Column("id").
			Where("abi = ?abi").
			Where("codehash = ?codehash").
			OnConflict("DO NOTHING").
			Returning("id").
			SelectOrInsert()
		if err != nil {
			errMessage := "failed to select or insert artifact"
			agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
			return errors.FromError(err).SetMessage(errMessage)
		}

		// Insert tag
		tag := models.NewTag(contract.Tag, repository.ID, artifact.ID)
		err = c.ModelContext(ctx, tag).
			OnConflict("ON CONSTRAINT tags_name_repository_id_key DO UPDATE").
			Set("artifact_id = ?artifact_id").
			Insert()
		if err != nil {
			errMessage := "failed to insert tag"
			agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
			return errors.FromError(err).SetMessage(errMessage)
		}

		return nil
	})
}

func (agent *PGContract) RegisterDeployment(ctx context.Context, chainID string, address ethcommon.Address, codeHash []byte) error {
	codehash := models.NewCodeHash(chainID, address, codeHash)

	// If uniqueness constraint is broken then it updates the former value
	err := agent.client.ModelContext(ctx, codehash).
		OnConflict("ON CONSTRAINT codehashes_chain_id_address_key DO UPDATE").
		Set("chain_id = ?chain_id").
		Set("address = ?address").
		Set("codehash = ?codehash").
		Returning("*").
		Insert()
	if err != nil {
		errMessage := "could not register contract deployment"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
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
	err := agent.client.QueryOneContext(ctx, qContract, query, codeHash)
	if err != nil {
		errMessage := "could not find contract by codehash"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
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
	err := agent.client.QueryOneContext(ctx, qContract, query, address)
	if err != nil {
		errMessage := "could not find contract by address"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return parseContract(qContract)
}

func (agent *PGContract) FindOneByNameAndTag(ctx context.Context, name, tag string) (*entities.Contract, error) {
	artifact := &models.Artifact{}
	err := agent.client.ModelContext(ctx, artifact).
		Column("artifact.id", "abi", "bytecode", "deployed_bytecode").
		Join("JOIN tags AS t ON t.artifact_id = artifact.id").
		Join("JOIN repositories AS registry ON registry.id = t.repository_id").
		Where("LOWER(t.name) = LOWER(?)", tag).
		Where("LOWER(registry.name) = LOWER(?)", name).
		SelectOne()
	if err != nil {
		errMessage := "could not find contract by name and tag"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return artifact.ToContract(name, tag), nil
}

func (agent *PGContract) ListNames(ctx context.Context) ([]string, error) {
	var names []string
	err := agent.client.ModelContext(ctx, (*models.Repository)(nil)).
		Column("name").
		OrderExpr("lower(name)").
		SelectColumn(&names)
	if err != nil && !errors.IsNotFoundError(err) {
		errMessage := "failed to fetch repository names"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return names, nil
}

func (agent *PGContract) ListTags(ctx context.Context, name string) ([]string, error) {
	var tags []string
	err := agent.client.ModelContext(ctx, (*models.Tag)(nil)).
		Column("tag.name").
		Join("JOIN repositories AS registry ON registry.id = tag.repository_id").
		Where("lower(registry.name) = lower(?)", name).
		OrderExpr("lower(tag.name)").
		SelectColumn(&tags)
	if err != nil {
		errMessage := "failed to fetch tags"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return tags, nil
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
