package pushnotification

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=push_notification.go -destination=mocks/push_notification.go -package=mocks

type Notifier interface {
	SendTxResponse(ctx context.Context, eventStream *entities.EventStream, job *entities.Job, errStr string) error
}
