package messenger

import (
	"context"

	"github.com/consensys/orchestrate/src/infra/messenger/types"
)

//go:generate mockgen -source=consumer.go -destination=mocks/consumer.go -package=mocks

type Consumer interface {
	Consume(ctx context.Context) error
	AppendHandler(msgType types.ConsumerRequestMessageType, msgHandler types.MessageHandler)
	Checker() error
	Close() error
}
