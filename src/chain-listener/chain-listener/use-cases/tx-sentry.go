package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=tx-listener.go -destination=mocks/tx-listener.go -package=mocks

type RetryJobSessionManager interface {
	StartSession(ctx context.Context, job *entities.Job) error
	StopSession(ctx context.Context, sessID string) error
}

type SendRetryJob interface {
	Execute(ctx context.Context, job *entities.Job, lastChildUUID string, nChildren int) (string, error)
}
