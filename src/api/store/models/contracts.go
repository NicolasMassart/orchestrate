package models

import (
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Artifact struct {
	tableName struct{} `pg:"artifacts"` // nolint:unused,structcheck // reason

	// UUID technical identifier
	ID int

	// Artifact data
	ABI              string `pg:"alias:abi"`
	Bytecode         string
	DeployedBytecode string
	// Codehash stored on the Ethereum account. Correspond to the hash of the deployedBytecode
	Codehash string
}

func NewArtifact(contract *entities.Contract) *Artifact {
	return &Artifact{
		ABI:              contract.RawABI,
		Bytecode:         contract.Bytecode.String(),
		DeployedBytecode: contract.DeployedBytecode.String(),
		Codehash:         contract.CodeHash.String(),
	}
}

func (a *Artifact) ToContract(name, tag string) *entities.Contract {
	return &entities.Contract{
		Name:             name,
		Tag:              tag,
		RawABI:           a.ABI,
		Bytecode:         hexutil.MustDecode(a.Bytecode),
		DeployedBytecode: hexutil.MustDecode(a.DeployedBytecode),
	}
}

type Codehash struct {
	tableName struct{} `pg:"codehashes"` // nolint:unused,structcheck // reason

	// UUID technical identifier
	ID int

	// Artifact data
	ChainID  string `pg:"alias:chain_id"`
	Address  string
	Codehash string
}

func NewCodeHash(chainID string, address common.Address, codeHash []byte) *Codehash {
	return &Codehash{
		ChainID:  chainID,
		Address:  address.Hex(),
		Codehash: hexutil.Encode(codeHash),
	}
}

type Event struct {
	tableName struct{} `pg:"events"` // nolint:unused,structcheck // reason

	// UUID technical identifier
	ID int

	// Artifact data
	Codehash          string
	SigHash           string
	IndexedInputCount uint `pg:",use_zero"`

	ABI string
}

func NewEvent(event *entities.ContractEvent) *Event {
	return &Event{
		ABI:               event.ABI,
		SigHash:           event.SigHash.String(),
		IndexedInputCount: event.IndexedInputCount,
		Codehash:          event.CodeHash.String(),
	}
}

func NewEvents(events []Event) []*entities.ContractEvent {
	res := []*entities.ContractEvent{}
	for _, e := range events {
		res = append(res, e.ToEntity())
	}
	return res
}

func (event *Event) ToEntity() *entities.ContractEvent {
	res := &entities.ContractEvent{
		ABI:               event.ABI,
		IndexedInputCount: event.IndexedInputCount,
	}

	if event.SigHash != "" {
		res.SigHash = hexutil.MustDecode(event.SigHash)
	}

	if event.Codehash != "" {
		res.CodeHash = hexutil.MustDecode(event.Codehash)
	}

	return res
}

type Repository struct {
	tableName struct{} `pg:"repositories"` // nolint:unused,structcheck // reason

	// UUID technical identifier
	ID int

	// Repository name
	Name string
}

func NewRepository(name string) *Repository {
	return &Repository{Name: name}
}

type Tag struct {
	tableName struct{} `pg:"tags"` // nolint:unused,structcheck // reason

	// UUID technical identifier
	ID int

	// Tag name
	Name         string
	RepositoryID int

	ArtifactID int
}

func NewTag(name string, repositoryID, artifactID int) *Tag {
	return &Tag{
		Name:         name,
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
	}
}
