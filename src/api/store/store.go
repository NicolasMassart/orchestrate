package store

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/consensys/orchestrate/src/infra/database"
)

//go:generate mockgen -source=store.go -destination=mocks/mock.go -package=mocks

type Store interface {
	Connect(ctx context.Context, conf interface{}) (DB, error)
}

type Agents interface {
	Schedule() ScheduleAgent
	Job() JobAgent
	Log() LogAgent
	Transaction() TransactionAgent
	TransactionRequest() TransactionRequestAgent
	Account() AccountAgent
	Faucet() FaucetAgent
	ContractEvent() ContractEventAgent
	Contract() ContractAgent
	Chain() ChainAgent
}

type DB interface {
	database.DB
	Agents
}

type Tx interface {
	database.Tx
	Agents
}

type TransactionRequestAgent interface {
	Insert(ctx context.Context, txRequest *entities.TxRequest, requestHash string, scheduleUUID string) error
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
	Insert(ctx context.Context, job *entities.Job, scheduleUUID, txUUID string) error
	Update(ctx context.Context, job *entities.Job) error
	FindOneByUUID(ctx context.Context, uuid string, tenants []string, ownerID string, withLogs bool) (*entities.Job, error)
	LockOneByUUID(ctx context.Context, uuid string) error
	Search(ctx context.Context, filters *entities.JobFilters, tenants []string, ownerID string) ([]*entities.Job, error)
}

type LogAgent interface {
	Insert(ctx context.Context, log *entities.Log, jobUUID string) error
}

type TransactionAgent interface {
	Insert(ctx context.Context, tx *entities.ETHTransaction) error
	Update(ctx context.Context, tx *entities.ETHTransaction, jobUUID string) error
	FindOneByJobUUID(ctx context.Context, jobUUID string) (*entities.ETHTransaction, error)
}
type AccountAgent interface {
	Insert(ctx context.Context, account *entities.Account) error
	Update(ctx context.Context, account *entities.Account) error
	FindOneByAddress(ctx context.Context, address string, tenants []string, ownerID string) (*entities.Account, error)
	Search(ctx context.Context, filters *entities.AccountFilters, tenants []string, ownerID string) ([]*entities.Account, error)
}

type FaucetAgent interface {
	Insert(ctx context.Context, faucet *entities.Faucet) error
	Update(ctx context.Context, faucet *entities.Faucet, tenants []string) error
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

type PrivateTxManagerAgent interface {
	Insert(ctx context.Context, privateTxManager *entities.PrivateTxManager) error
	Update(ctx context.Context, privateTxManager *entities.PrivateTxManager) error
	Search(ctx context.Context, chainUUID string) ([]*entities.PrivateTxManager, error)
	Delete(ctx context.Context, privateTxManagerUUID string) error
}

type ContractAgent interface {
	Register(ctx context.Context, contract *entities.Contract) error
	RegisterDeployment(ctx context.Context, chainID string, address ethcommon.Address, codeHash hexutil.Bytes) error
	FindOneByNameAndTag(ctx context.Context, name, tag string) (*entities.Contract, error)
	FindOneByCodeHash(ctx context.Context, codeHash string) (*entities.Contract, error)
	FindOneByAddress(ctx context.Context, address string) (*entities.Contract, error)
	ListNames(ctx context.Context) ([]string, error)
	ListTags(ctx context.Context, name string) ([]string, error)
}

type ContractEventAgent interface {
	RegisterMultiple(ctx context.Context, events []*entities.ContractEvent) error
	FindOneByAccountAndSigHash(ctx context.Context, chainID, address, sighash string, indexedInputCount uint32) (*entities.ContractEvent, error)
	FindDefaultBySigHash(ctx context.Context, sighash string, indexedInputCount uint32) ([]*entities.ContractEvent, error)
}
