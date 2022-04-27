package sessions

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=sessions.go -destination=mocks/sessions.go -package=mocks

type ChainSessionManager interface {
	StartSession(ctx context.Context, chainUUID string) error
}

type RetryJobSessionManager interface {
	StartSession(ctx context.Context, job *entities.Job) error
}

type SessionManagers interface {
	ChainSessionManager() ChainSessionManager
	RetryJobSessionManager() RetryJobSessionManager
}
