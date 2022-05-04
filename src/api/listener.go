package api

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/messenger"
)

const apiConsumerComponent = "api.service.consumer"

type ConsumerSrv struct {
	consumer messenger.Consumer
	logger   *log.Logger
}

func NewConsumerService(consumer messenger.Consumer) *ConsumerSrv {
	return &ConsumerSrv{
		consumer: consumer,
		logger:   log.NewLogger().SetComponent(apiConsumerComponent),
	}
}

func (s *ConsumerSrv) Run(ctx context.Context) error {
	s.logger.Debug("starting service...")
	err := s.consumer.Consume(ctx)
	if err != nil {
		// @TODO Exit gracefully in case of context canceled
		s.logger.WithError(err).Error("service exited with errors")
	}

	return err
}

func (s *ConsumerSrv) Close() error {
	return s.consumer.Close()
}
