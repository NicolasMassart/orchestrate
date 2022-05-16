package messenger

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=consumer.go -destination=mocks/consumer.go -package=mocks

type Consumer interface {
	Consume(ctx context.Context) error
	AppendHandler(msgType entities.RequestMessageType, msgHandler MessageHandler)
	Checker() error
	Close() error
}

type MessageHandler func(ctx context.Context, msg *entities.Message) error
