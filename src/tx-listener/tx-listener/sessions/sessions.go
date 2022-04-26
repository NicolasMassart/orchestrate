package sessions

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=sessions.go -destination=mocks/sessions.go -package=mocks

type ChainSessionManager interface {
	StartSession(ctx context.Context, chainUUID string) error
}

type TxSentrySessionManager interface {
	StartSession(ctx context.Context, job *entities.Job) error
	StopSession(ctx context.Context, sessID string) error
}

type SessionManagers interface {
	ChainSessionManager() ChainSessionManager
	TxSentrySessionManager() TxSentrySessionManager
}
