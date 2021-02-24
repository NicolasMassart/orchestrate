package store

import (
	"context"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/entities"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/database"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/store/models"
	"google.golang.org/grpc"
)

//go:generate mockgen -source=store.go -destination=mocks/mock.go -package=mocks

type Builder interface {
	Build(ctx context.Context, name string, configuration interface{}) (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor, func(srv *grpc.Server), error)
}

type Store interface {
	Connect(ctx context.Context, conf interface{}) (DB, error)
}

type Agents interface {
	Schedule() ScheduleAgent
	Job() JobAgent
	Log() LogAgent
	Transaction() TransactionAgent
	TransactionRequest() TransactionRequestAgent
}

type DB interface {
	database.DB
	Agents
}

type Tx interface {
	database.Tx
	Agents
}

// Interfaces data agents
type TransactionRequestAgent interface {
	Insert(ctx context.Context, txRequest *models.TransactionRequest) error
	FindOneByIdempotencyKey(ctx context.Context, idempotencyKey string, tenantID string) (*models.TransactionRequest, error)
	FindOneByUUID(ctx context.Context, scheduleUUID string, tenants []string) (*models.TransactionRequest, error)
	Search(ctx context.Context, filters *entities.TransactionFilters, tenants []string) ([]*models.TransactionRequest, error)
}

type ScheduleAgent interface {
	Insert(ctx context.Context, schedule *models.Schedule) error
	FindOneByUUID(ctx context.Context, uuid string, tenants []string) (*models.Schedule, error)
	FindAll(ctx context.Context, tenants []string) ([]*models.Schedule, error)
}

type JobAgent interface {
	Insert(ctx context.Context, job *models.Job) error
	Update(ctx context.Context, job *models.Job) error
	FindOneByUUID(ctx context.Context, uuid string, tenants []string, withLogs bool) (*models.Job, error)
	LockOneByUUID(ctx context.Context, uuid string) error
	Search(ctx context.Context, filters *entities.JobFilters, tenants []string) ([]*models.Job, error)
}

type LogAgent interface {
	Insert(ctx context.Context, log *models.Log) error
}

type TransactionAgent interface {
	Insert(ctx context.Context, tx *models.Transaction) error
	Update(ctx context.Context, tx *models.Transaction) error
}
