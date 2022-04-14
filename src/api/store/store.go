package store

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=store.go -destination=mocks/mock.go -package=mocks

type DB interface {
	Schedule() ScheduleAgent
	Job() JobAgent
	TransactionRequest() TransactionRequestAgent
	Account() AccountAgent
	Faucet() FaucetAgent
	ContractEvent() ContractEventAgent
	Contract() ContractAgent
	Chain() ChainAgent
	EventStream() EventStreamAgent
	RunInTransaction(ctx context.Context, persistFunc func(db DB) error) error
}

type TransactionRequestAgent interface {
	Insert(ctx context.Context, txRequest *entities.TxRequest, requestHash string, scheduleUUID string) (*entities.TxRequest, error)
	FindOneByIdempotencyKey(ctx context.Context, idempotencyKey string, tenantID string, ownerID string) (*entities.TxRequest, error)
	FindOneByUUID(ctx context.Context, scheduleUUID string, tenants []string, ownerID string) (*entities.TxRequest, error)
	Search(ctx context.Context, filters *entities.TransactionRequestFilters, tenants []string, ownerID string) ([]*entities.TxRequest, error)
}

type ScheduleAgent interface {
	Insert(ctx context.Context, schedule *entities.Schedule) error
	FindOneByUUID(ctx context.Context, uuid string, tenants []string, ownerID string) (*entities.Schedule, error)
	FindAll(ctx context.Context, tenants []string, ownerID string) ([]*entities.Schedule, error)
}

type JobAgent interface {
	Insert(ctx context.Context, job *entities.Job, log *entities.Log) error
	Update(ctx context.Context, job *entities.Job, log *entities.Log) error
	FindOneByUUID(ctx context.Context, uuid string, tenants []string, ownerID string, withLogs bool) (*entities.Job, error)
	Search(ctx context.Context, filters *entities.JobFilters, tenants []string, ownerID string) ([]*entities.Job, error)
	GetSiblingJobs(ctx context.Context, parentJobUUID string, tenants []string, ownerID string) ([]*entities.Job, error)
}

type AccountAgent interface {
	Insert(ctx context.Context, account *entities.Account) (*entities.Account, error)
	Update(ctx context.Context, account *entities.Account) (*entities.Account, error)
	FindOneByAddress(ctx context.Context, address string, tenants []string, ownerID string) (*entities.Account, error)
	Search(ctx context.Context, filters *entities.AccountFilters, tenants []string, ownerID string) ([]*entities.Account, error)
}

type FaucetAgent interface {
	Insert(ctx context.Context, faucet *entities.Faucet) (*entities.Faucet, error)
	Update(ctx context.Context, faucet *entities.Faucet, tenants []string) (*entities.Faucet, error)
	FindOneByUUID(ctx context.Context, uuid string, tenants []string) (*entities.Faucet, error)
	Search(ctx context.Context, filters *entities.FaucetFilters, tenants []string) ([]*entities.Faucet, error)
	Delete(ctx context.Context, faucet *entities.Faucet, tenants []string) error
}

type ChainAgent interface {
	Insert(ctx context.Context, chain *entities.Chain) error
	Update(ctx context.Context, chain *entities.Chain, tenants []string, ownerID string) error
	Search(ctx context.Context, filters *entities.ChainFilters, tenants []string, ownerID string) ([]*entities.Chain, error)
	FindOneByUUID(ctx context.Context, uuid string, tenants []string, ownerID string) (*entities.Chain, error)
	FindOneByName(ctx context.Context, name string, tenants []string, ownerID string) (*entities.Chain, error)
	Delete(ctx context.Context, chain *entities.Chain, tenants []string) error
}

type ContractAgent interface {
	Register(ctx context.Context, contract *entities.Contract) error
	Update(ctx context.Context, contract *entities.Contract) error
	RegisterDeployment(ctx context.Context, chainID string, address ethcommon.Address, codeHash []byte) error
	FindOneByNameAndTag(ctx context.Context, name, tag string) (*entities.Contract, error)
	FindOneByCodeHash(ctx context.Context, codeHash string) (*entities.Contract, error)
	FindOneByAddress(ctx context.Context, address string) (*entities.Contract, error)
	ListNames(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context, name string) ([]string, error)
}

type ContractEventAgent interface {
	RegisterMultiple(ctx context.Context, events []entities.ContractEvent) error
	FindOneByAccountAndSigHash(ctx context.Context, chainID, address, sighash string, indexedInputCount uint32) (*entities.ContractEvent, error)
	FindDefaultBySigHash(ctx context.Context, sighash string, indexedInputCount uint32) ([]*entities.ContractEvent, error)
}

type EventStreamAgent interface {
	Insert(ctx context.Context, eventStream *entities.EventStream) (*entities.EventStream, error)
	Search(ctx context.Context, filters *entities.EventStreamFilters, tenants []string, ownerID string) ([]*entities.EventStream, error)
	FindOneByUUID(ctx context.Context, uuid string, tenants []string, ownerID string) (*entities.EventStream, error)
	FindOneByTenantAndChain(ctx context.Context, tenantID, chainUUID string, tenants []string, ownerID string) (*entities.EventStream, error)
	Delete(ctx context.Context, uuid string, tenants []string, ownerID string) error
	Update(ctx context.Context, eventStream *entities.EventStream, tenants []string, ownerID string) (*entities.EventStream, error)
}
