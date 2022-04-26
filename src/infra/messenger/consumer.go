package messenger

import (
	"context"
)

//go:generate mockgen -source=consumer.go -destination=mocks/consumer.go -package=mocks

type Consumer interface {
	Consume(ctx context.Context) error
	Checker() error
	Close() error
}
