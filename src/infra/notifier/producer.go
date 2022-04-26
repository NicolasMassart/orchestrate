package notifier

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=producer.go -destination=mocks/producer.go -package=mocks

type Producer interface {
	SendTxResponse(ctx context.Context, eventStream *entities.EventStream, job *entities.Job, errStr string) error
}
