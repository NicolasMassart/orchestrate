package store

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source=store.go -destination=mocks/store.go -package=mocks

type Chain interface {
	Add(ctx context.Context, chain *entities.Chain) error
	Update(ctx context.Context, chain *entities.Chain) error
	Delete(ctx context.Context, chainUUID string) error
	Get(ctx context.Context, chainUUID string) (*entities.Chain, error)
}

type PendingJob interface {
	Add(ctx context.Context, job *entities.Job) error
	Remove(ctx context.Context, jobUUID string) error
	Update(ctx context.Context, job *entities.Job) error
	GetByTxHash(ctx context.Context, chainUUID string, txHash *ethcommon.Hash) (*entities.Job, error)
	GetJobUUID(ctx context.Context, jobUUID string) (*entities.Job, error)
	ListPerChainUUID(ctx context.Context, chainUUID string) ([]*entities.Job, error)
	DeletePerChainUUID(ctx context.Context, chainUUID string) error
}

type RetryJobSession interface {
	Add(ctx context.Context, job *entities.Job) error
	Has(ctx context.Context, jobUUID string) bool
	Remove(ctx context.Context, jobUUID string) error
	GetByTxHash(ctx context.Context, chainUUID string, txHash *ethcommon.Hash) (string, error)
	ListByChainUUID(ctx context.Context, chainUUID string) ([]string, error)
	DeletePerChainUUID(ctx context.Context, chainUUID string) error
}

type State interface {
	ChainState() Chain
	PendingJobState() PendingJob
	RetryJobSessionState() RetryJobSession
}
