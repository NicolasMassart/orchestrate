package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/src/notifier/service"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	api "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/infra/messenger/kafka"
	"github.com/consensys/orchestrate/src/notifier/notifier/builder"
	"github.com/hashicorp/go-multierror"
)

const component = "application.notifier"

type Daemon struct {
	consumers []messenger.Consumer
	config    *Config
	logger    *log.Logger
	cancel    context.CancelFunc
}

var _ app.Daemon = &Daemon{}

func New(config *Config, eventStreamClient api.EventStreamClient, kafkaNotifier, webhookNotifier messenger.Producer) (*Daemon, error) {
	// Create business layer use cases
	useCases := builder.NewUseCases(kafkaNotifier, webhookNotifier)
	msgConsumerHandler := service.NewMessageConsumerHandler(useCases, eventStreamClient, config.MaxRetries)

	consumers := make([]messenger.Consumer, config.NConsumer)
	for idx := 0; idx < config.NConsumer; idx++ {
		var err error
		consumers[idx], err = kafka.NewMessageConsumer(config.Kafka, []string{config.ConsumerTopic}, msgConsumerHandler)
		if err != nil {
			return nil, err
		}
	}

	return &Daemon{
		consumers: consumers,
		config:    config,
		logger:    log.NewLogger().SetComponent(component),
	}, nil
}

func (d *Daemon) Run(ctx context.Context) error {
	d.logger.Debug("starting notifier")

	ctx, d.cancel = context.WithCancel(ctx)
	ctx = multitenancy.WithUserInfo(ctx, multitenancy.NewInternalAdminUser())

	gr := &multierror.Group{}
	for idx, consumerGroup := range d.consumers {
		cGroup := consumerGroup
		cGroupID := fmt.Sprintf("c-%d", idx)
		logger := d.logger.WithField("consumer", cGroupID)
		cctx := log.With(log.WithField(ctx, "consumer", cGroupID), logger)

		gr.Go(func() error {
			// We retry once after consume exits to prevent entire stack to exit after kafka rebalance is triggered
			err := backoff.RetryNotify(
				func() error {
					err := cGroup.Consume(cctx)

					// In this case, kafka rebalance was triggered and we want to retry
					if err == nil && cctx.Err() == nil {
						return fmt.Errorf("kafka rebalance was triggered")
					}

					return backoff.Permanent(err)
				},
				backoff.NewConstantBackOff(time.Millisecond*500),
				func(err error, duration time.Duration) {
					logger.WithError(err).Warnf("consuming session exited, retrying in %s", duration.String())
				},
			)
			d.cancel()
			return err
		})
	}

	return gr.Wait().ErrorOrNil()
}

func (d *Daemon) Close() error {
	var gerr error
	for _, consumerGroup := range d.consumers {
		gerr = errors.CombineErrors(gerr, consumerGroup.Close())
	}

	return gerr
}
