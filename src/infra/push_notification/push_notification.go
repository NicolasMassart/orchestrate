package pushnotification

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=notifier.go -destination=mocks/notifier.go -package=mocks

type Notifier interface {
	SendTxResponse(ctx context.Context, eventStream *entities.EventStream, job *entities.Job, errStr string) error
}
