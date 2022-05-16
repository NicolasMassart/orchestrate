package integrationtests

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/messenger"

	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/notifier"
	"github.com/consensys/orchestrate/tests/pkg/trackers"

	"github.com/consensys/orchestrate/cmd/flags"
	webhook "github.com/consensys/orchestrate/src/infra/webhook/http"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	httputils "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/tests/pkg/docker"
	"github.com/consensys/orchestrate/tests/pkg/docker/config"
	kafkaDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/kafka"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/zookeeper"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/traefik/traefik/v2/pkg/log"
	"gopkg.in/h2non/gock.v1"
)

const kafkaContainerID = "notifier-kafka-api"
const zookeeperContainerID = "notifier-zookeeper-api"
const networkName = "notifier"

var envKafkaHostPort string

type IntegrationEnvironment struct {
	ctx                      context.Context
	logger                   log.Logger
	notifier                 app.Daemon
	client                   *docker.Client
	messengerConsumerTracker *trackers.MessengerConsumerTracker
	notifierConsumerTracker  *trackers.NotifierConsumerTracker
	notificationTopic        string
	cfg                      *notifier.Config
}

func NewIntegrationEnvironment(ctx context.Context) (*IntegrationEnvironment, error) {
	logger := log.FromContext(ctx)
	envKafkaHostPort = strconv.Itoa(utils.RandIntRange(20000, 29092))

	apiTopic := "topic-api"
	notifierTopic := "topic-notifier"
	decodedTopic := "topic-decoded"

	// Define external hostname
	kafkaExternalHostname := fmt.Sprintf("localhost:%s", envKafkaHostPort)

	// Initialize environment flags
	flgs := pflag.NewFlagSet("api-integration-test", pflag.ContinueOnError)
	flags.NewAPIFlags(flgs)
	flags.NotifierFlags(flgs)
	args := []string{
		"--kafka-url=" + kafkaExternalHostname,
		"--topic-api=" + apiTopic,
		"--topic-notifier=" + notifierTopic,
		"--log-level=panic",
	}

	err := flgs.Parse(args)
	if err != nil {
		logger.WithError(err).Error("cannot parse environment flags")
		return nil, err
	}

	// Initialize environment container setup
	composition := &config.Composition{
		Containers: map[string]*config.Container{
			zookeeperContainerID: {Zookeeper: zookeeper.NewDefault()},
			kafkaContainerID: {Kafka: kafkaDocker.NewDefault().
				SetHostPort(envKafkaHostPort).
				SetZookeeperHostname(zookeeperContainerID).
				SetKafkaInternalHostname(kafkaContainerID).
				SetKafkaExternalHostname(kafkaExternalHostname),
			},
		},
	}

	// Docker client
	dockerClient, err := docker.NewClient(composition)
	if err != nil {
		logger.WithError(err).Error("cannot initialize new environment")
		return nil, err
	}

	return &IntegrationEnvironment{
		ctx:               ctx,
		logger:            logger,
		client:            dockerClient,
		cfg:               flags.NewNotifierConfig(viper.GetViper()),
		notificationTopic: decodedTopic,
	}, nil
}

func (env *IntegrationEnvironment) Start(ctx context.Context) error {
	err := env.client.CreateNetwork(ctx, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not create network")
		return err
	}

	// Start Kafka + zookeeper
	err = env.client.Up(ctx, zookeeperContainerID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up zookeeper")
		return err
	}

	err = env.client.Up(ctx, kafkaContainerID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up Kafka")
		return err
	}
	err = env.client.WaitTillIsReady(ctx, kafkaContainerID, 20*time.Second)
	if err != nil {
		env.logger.WithError(err).Error("could not start Kafka")
		return err
	}

	env.notifier, err = newNotifier(env.cfg)
	if err != nil {
		env.logger.WithError(err).Error("cannot initialize notifier daemon")
		return err
	}

	// Start internal kafka consumer
	env.messengerConsumerTracker, err = trackers.NewMessengerConsumerTracker(*env.cfg.Kafka, []string{env.cfg.Messenger.TopicAPI})
	if err != nil {
		env.logger.WithError(err).Error("could initialize kafka internal Consumer")
		return err
	}

	// Start internal kafka consumer
	env.notifierConsumerTracker, err = trackers.NewNotifierConsumerTacker(*env.cfg.Kafka, []string{env.notificationTopic})
	if err != nil {
		env.logger.WithError(err).Error("could initialize kafka internal Consumer")
		return err
	}

	go func() {
		err := env.messengerConsumerTracker.Consume(ctx)
		if err != nil {
			env.logger.WithError(err).Error("could not start API")
		}
	}()

	go func() {
		err := env.notifierConsumerTracker.Consume(ctx)
		if err != nil {
			env.logger.WithError(err).Error("could not start API")
		}
	}()

	// Start API
	go func() {
		err := env.notifier.Run(ctx)
		if err != nil {
			env.logger.WithError(err).Error("could not start API")
		}
	}()

	// Wait for consumers to be ready
	time.Sleep(time.Second * 10)

	return nil
}

func (env *IntegrationEnvironment) Teardown(ctx context.Context) {
	env.logger.Info("tearing test suite down")

	if env.notifier != nil {
		err := env.notifier.Close()
		if err != nil {
			env.logger.WithError(err).Error("could not stop API")
		}
	}

	err := env.client.Down(ctx, kafkaContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down Kafka")
	}

	err = env.client.Down(ctx, zookeeperContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down zookeeper")
	}

	err = env.client.RemoveNetwork(ctx, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not remove network")
	}
}

func newNotifier(cfg *notifier.Config) (app.Daemon, error) {
	kafkaProducer, err := sarama.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, err
	}

	messengerClient := messenger.NewProducerClient(&messenger.Config{
		TopicAPI:      cfg.Messenger.TopicAPI,
		TopicNotifier: cfg.ConsumerTopic,
	}, kafkaProducer)

	interceptedHTTPClient := httputils.NewClient(httputils.NewDefaultConfig())
	gock.InterceptClient(interceptedHTTPClient)
	webhookProducer := webhook.New(interceptedHTTPClient)

	return notifier.New(cfg, kafkaProducer, webhookProducer, messengerClient)
}
