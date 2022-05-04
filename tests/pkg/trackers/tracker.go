package trackers

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
)

type ConsumerTracker struct {
	chanRegistry *utils.ChanRegistry
	logger       *log.Logger
}

func NewConsumerTracker(chanRegistry *utils.ChanRegistry, logger *log.Logger) *ConsumerTracker {
	return &ConsumerTracker{
		chanRegistry: chanRegistry,
		logger:       logger,
	}
}

func (m *ConsumerTracker) WaitForMessage(ctx context.Context, msgId string, timeout time.Duration) (interface{}, error) {
	logger := m.logger.WithField("id", msgId).WithField("timeout", timeout/time.Millisecond)

	logger.Debug("waiting for message...")
	msg, err := m.waitForChanMessage(ctx, msgId, timeout)
	if err != nil {
		logger.WithError(err).Error("failed to find message")
		return nil, err
	}

	return msg, err
}

func (m *ConsumerTracker) waitForChanMessage(ctx context.Context, msgId string, timeout time.Duration) (interface{}, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var ch = make(chan interface{}, 1)
	go func(chx chan interface{}) {
		if !m.chanRegistry.HasChan(msgId) {
			m.chanRegistry.Register(msgId, make(chan interface{}, 1))
		}

		e := <-m.chanRegistry.GetChan(msgId)
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

func keyGenOf(msgId, msgType string) string {
	return fmt.Sprintf("%s/%s", msgType, msgId)
}
