package testutils

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
)

type ConsumerTracker struct {
	chanRegistry *utils.ChanRegistry
	logger       *log.Logger
	keyGenOf     func(key, topic string) string
}

func NewConsumerTracker(chanRegistry *utils.ChanRegistry, keyGenOf func(key, topic string) string, logger *log.Logger) *ConsumerTracker {
	return &ConsumerTracker{
		chanRegistry: chanRegistry,
		logger:       logger,
		keyGenOf:     keyGenOf,
	}
}

func (m *ConsumerTracker) WaitForMessage(ctx context.Context, id, topic string, timeout time.Duration) (interface{}, error) {
	logger := m.logger.WithField("id", id).WithField("topic", topic).WithField("timeout", timeout/time.Millisecond)

	logger.Debug("waiting for message...")
	msg, err := m.waitForChanMessage(ctx, id, topic, timeout)
	if err != nil {
		logger.WithError(err).Error("failed to find message")
		return nil, err
	}

	return msg, err
}

func (m *ConsumerTracker) waitForChanMessage(ctx context.Context, id, topic string, timeout time.Duration) (interface{}, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var ch = make(chan interface{}, 1)
	go func(chx chan interface{}) {
		msgKey := m.keyGenOf(id, topic)
		if !m.chanRegistry.HasChan(msgKey) {
			m.chanRegistry.Register(msgKey, make(chan interface{}, 1))
		}

		e := <-m.chanRegistry.GetChan(msgKey)
		chx <- e
	}(ch)

	select {
	case e := <-ch:
		return e, nil
	case <-cctx.Done():
		return nil, cctx.Err()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
