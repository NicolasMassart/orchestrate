package listener

import (
	"context"

	"github.com/consensys/orchestrate/src/chain-listener/listener/events"
)

//go:generate mockgen -source=listener.go -destination=mocks/listener.go -package=mocks

type ChainListener interface {
	Run(context.Context) error
	Close() error
	Subscribe(chan *events.Chain) string
	Unsubscribe(string) error
}

type ChainBlockListener interface {
	Run(ctx context.Context) error
	Close() error
}

type ChainPendingJobsListener interface {
	Run(context.Context) error
	Close() error
}
