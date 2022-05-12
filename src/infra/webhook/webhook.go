package webhook

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=webhook.go -destination=mocks/webhook.go -package=mocks

type Producer interface {
	Send(ctx context.Context, webhook *entities.EventStreamWebhookSpec, body interface{}) error
}
