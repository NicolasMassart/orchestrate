package e2e

import (
	"context"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	loader "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/handlers/loader/sarama"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/handlers/offset"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/app"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/app/worker"
	broker "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/broker/sarama"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/engine"
	orchlog "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/logger"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/utils"
	contractregistry "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/contract-registry/client"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/tests/handlers"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/tests/handlers/dispatcher"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/tests/service/cucumber"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/tests/service/cucumber/steps"
)

var (
	appli       *app.App
	startOnce   = &sync.Once{}
	readyToTest chan bool
	done        chan struct{}
	cancel      func()
)

func LongKeyOf(topics map[string]string) dispatcher.KeyOfFunc {
	return func(txctx *engine.TxContext) (string, error) {
		topic, ok := topics[txctx.In.Entrypoint()]
		if !ok {
			return "", fmt.Errorf("unknown message entrypoint")
		}

		scenario := txctx.Envelope.GetContextLabelsValue("scenario.id")
		if scenario == "" {
			return "", fmt.Errorf("message has no test scenario")
		}

		return steps.LongKeyOf(topic, scenario, txctx.Envelope.GetID()), nil
	}
}

func ShortKeyOf(topics map[string]string) dispatcher.KeyOfFunc {
	return func(txtcx *engine.TxContext) (string, error) {
		topic, ok := topics[txtcx.In.Entrypoint()]
		if !ok {
			return "", fmt.Errorf("unknown message entrypoint")
		}

		scenario := txtcx.Envelope.GetContextLabelsValue("scenario.id")
		if scenario == "" {
			return "", fmt.Errorf("message has no test scenario")
		}

		return steps.ShortKeyOf(topic, scenario), nil
	}
}

func initComponents(ctx context.Context) {
	utils.InParallel(
		// Initialize Engine
		func() {
			engine.Init(ctx)
		},
		// Initialize ConsumerGroup
		func() {
			viper.Set(broker.KafkaGroupViperKey, "group-e2e")
			broker.InitConsumerGroup(ctx)
		},
		// Initialize Handlers
		func() {
			// Prepare topics map for dispatcher
			topics := make(map[string]string)
			for _, topic := range steps.TOPICS {
				topics[viper.GetString(fmt.Sprintf("topic.%v", topic))] = topic
			}
			dispatcher.SetKeyOfFuncs(
				LongKeyOf(topics),
				ShortKeyOf(topics),
			)
			handlers.Init(ctx)
		},
		// Initialize cucumber handlers
		func() {
			cucumber.Init(ctx)
		},
		// Initialize the contractregistryClient
		func() {
			contractregistry.Init(ctx, viper.GetString(contractregistry.ContractRegistryURLViperKey))
		},
	)
}

func registerHandlers() {
	// Generic handlers on every worker
	engine.Register(loader.Loader)
	engine.Register(offset.Marker)
	engine.Register(dispatcher.GlobalHandler())
}

// Start starts application
func Start(ctx context.Context) error {
	startOnce.Do(func() {
		// Create context for application
		ctx, cancel = context.WithCancel(ctx)
		done = make(chan struct{})

		// Create Configuration
		cfg := app.NewConfig(viper.GetViper())
		orchlog.ConfigureLogger(cfg.HTTP)

		// Register all Handlers
		compCtx, cancelComponents := context.WithCancel(context.Background())
		initComponents(compCtx)
		registerHandlers()

		// Create appli to expose metrics
		appli = worker.New(cfg)
		_ = appli.Start(ctx)

		// Start consuming on every topics of interest
		var topics []string
		for _, topic := range steps.TOPICS {
			topics = append(topics, viper.GetString(fmt.Sprintf("topic.%v", topic)))
		}

		readyToTest = make(chan bool, 1)

		go func() {
			<-readyToTest
			cucumber.Run(func() {
				_ = Stop(context.Background())
			}, cucumber.GlobalOptions())
		}()

		cg := &EmbeddingConsumerGroupHandler{
			engine: broker.NewEngineConsumerGroupHandler(engine.GlobalEngine()),
		}

		go func() {
			log.FromContext(ctx).WithFields(logrus.Fields{
				"topics": topics,
			}).Info("connecting")

			err := broker.Consume(
				ctx,
				topics,
				cg,
			)
			if err != nil {
				log.FromContext(ctx).WithError(err).Error("error on consumer")
			}
			cancelComponents()
			close(done)
		}()
	})
	<-done
	return nil
}

type EmbeddingConsumerGroupHandler struct {
	engine *broker.EngineConsumerGroupHandler
}

func (h *EmbeddingConsumerGroupHandler) Setup(s sarama.ConsumerGroupSession) error {
	err := h.engine.Setup(s)
	readyToTest <- true
	return err
}

func (h *EmbeddingConsumerGroupHandler) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	return h.engine.ConsumeClaim(s, c)
}

func (h *EmbeddingConsumerGroupHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	return h.engine.Cleanup(s)
}

func Stop(ctx context.Context) error {
	cancel()
	_ = appli.Stop(ctx)
	<-done
	return nil
}