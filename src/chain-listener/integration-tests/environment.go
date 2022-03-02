package integrationtests

import (
	"context"
	"fmt"
	http2 "net/http"
	"strconv"
	"time"

	sarama2 "github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/cmd/flags"
	integrationtest "github.com/consensys/orchestrate/pkg/integration-test"
	"github.com/consensys/orchestrate/pkg/integration-test/docker"
	"github.com/consensys/orchestrate/pkg/integration-test/docker/config"
	ganacheDocker "github.com/consensys/orchestrate/pkg/integration-test/docker/container/ganache"
	kafkaDocker "github.com/consensys/orchestrate/pkg/integration-test/docker/container/kafka"
	"github.com/consensys/orchestrate/pkg/integration-test/docker/container/zookeeper"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	httputils "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
	chainlistener "github.com/consensys/orchestrate/src/chain-listener"
	listenermetrics "github.com/consensys/orchestrate/src/chain-listener/chain-listener/metrics"
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/h2non/gock.v1"
)

const kafkaContainerID = "chain-listener-kafka"
const zookeeperContainerID = "zookeeper-chain-listener"
const apiURL = "http://api:8081"
const apiMetricsURL = "http://api:8082"
const networkName = "chain-listener"
const maxRecoveryDefault = 1
const ganacheContainerID = "chain-listener-ganache-api"

var envKafkaHostPort string
var envMetricsPort string
var envGanacheHostPort string

type IntegrationEnvironment struct {
	ctx               context.Context
	cancel            context.CancelFunc
	logger            *log.Logger
	chainListener     *chainlistener.Service
	client            *docker.Client
	consumer          *integrationtest.KafkaConsumer
	producer          sarama2.SyncProducer
	metricsURL        string
	srvConfig         *chainlistener.Config
	blockchainNodeURL string
}

func NewIntegrationEnvironment(ctx context.Context, cancel context.CancelFunc) (*IntegrationEnvironment, error) {
	logger := log.NewLogger()
	envMetricsPort = strconv.Itoa(utils.RandIntRange(30000, 38082))
	envKafkaHostPort = strconv.Itoa(utils.RandIntRange(20000, 29092))
	envGanacheHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))

	// Define external hostname
	kafkaExternalHostname := fmt.Sprintf("localhost:%s", envKafkaHostPort)

	// Initialize environment flags
	flgs := pflag.NewFlagSet("chain-listener-integration-test", pflag.ContinueOnError)
	flags.ChainListenerFlags(flgs)
	args := []string{
		"--metrics-port=" + envMetricsPort,
		"--kafka-url=" + kafkaExternalHostname,
		"--api-url=" + apiURL,
		"--log-level=info",
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
			ganacheContainerID: {Ganache: ganacheDocker.NewDefault().SetHostPort(envGanacheHostPort)},
		},
	}

	dockerClient, err := docker.NewClient(composition)
	if err != nil {
		logger.WithError(err).Error("cannot initialize docker client")
		return nil, err
	}

	return &IntegrationEnvironment{
		ctx:               ctx,
		cancel:            cancel,
		logger:            logger,
		client:            dockerClient,
		srvConfig:         flags.NewChainListenerConfig(viper.GetViper()),
		metricsURL:        "http://localhost:" + envMetricsPort,
		producer:          sarama.GlobalSyncProducer(),
		blockchainNodeURL: fmt.Sprintf("http://localhost:%s", envGanacheHostPort),
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

	// Start ganache
	err = env.client.Up(ctx, ganacheContainerID, "")
	if err != nil {
		env.logger.WithError(err).Error("could not up ganache")
		return err
	}
	err = env.client.WaitTillIsReady(ctx, ganacheContainerID, 10*time.Second)
	if err != nil {
		env.logger.WithError(err).Error("could not start ganache")
		return err
	}

	// Create app
	env.chainListener, err = newChainListener(ctx, env.srvConfig)
	if err != nil {
		env.logger.WithError(err).Error("could not initialize chain-listener")
		return err
	}

	// Start Kafka consumer
	env.consumer, err = integrationtest.NewKafkaTestConsumer(
		ctx,
		"chain-listener-integration-listener-group",
		sarama.GlobalClient(),
		[]string{env.srvConfig.ChainListenerConfig.DecodedOutTopic},
	)
	if err != nil {
		env.logger.WithError(err).Error("could initialize Kafka")
		return err
	}
	err = env.consumer.Start(context.Background())
	if err != nil {
		env.logger.WithError(err).Error("could not run Kafka consumer")
		return err
	}

	// Set producer
	env.producer = sarama.GlobalSyncProducer()

	gock.New(apiURL).
		Get("/chains").
		Times(-1).
		Reply(http2.StatusOK).JSON([]api.ChainResponse{
		// {UUID: ganacheChainUUID},
	})

	// Start chain-listener app
	go func() {
		err = env.chainListener.Start(ctx)
		if err != nil {
			env.logger.WithError(err).Error("could not start chain-listener")
			env.cancel()
		}
	}()

	integrationtest.WaitForServiceLive(ctx, fmt.Sprintf("%s/live", env.metricsURL), "chain-listener", 15*time.Second)
	return nil
}

func (env *IntegrationEnvironment) Teardown(ctx context.Context) {
	env.logger.Info("tearing test suite down")

	err := env.chainListener.Stop(ctx)
	if err != nil {
		env.logger.WithError(err).Error("could not stop chain-listener")
	}

	err = env.client.Down(ctx, kafkaContainerID)
	if err != nil {
		env.logger.WithError(err).Errorf("could not down Kafka")
	}

	err = env.client.Down(ctx, zookeeperContainerID)
	if err != nil {
		env.logger.WithError(err).Errorf("could not down zookeeper")
	}

	err = env.client.Down(ctx, ganacheContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down ganache")
	}

	err = env.client.RemoveNetwork(ctx, networkName)
	if err != nil {
		env.logger.WithError(err).Errorf("could not remove network")
	}
}

func newChainListener(ctx context.Context, cfg *chainlistener.Config) (*chainlistener.Service, error) {
	// Initialize dependencies
	sarama.InitSyncProducer(ctx)
	sarama.InitConsumerGroup(ctx, cfg.ChainListenerConfig.DecodedOutTopic)

	httpClient := httputils.NewClient(httputils.NewDefaultConfig())
	gock.InterceptClient(httpClient)

	ec := ethclient.NewClient(testBackOff, httpClient)
	listenerMetrics := listenermetrics.NewListenerNopMetrics()

	conf2 := client.NewConfig(apiURL, "", nil)
	conf2.MetricsURL = apiMetricsURL
	apiClient := client.NewHTTPClient(httpClient, conf2)

	return chainlistener.NewChainListener(cfg, apiClient, sarama.GlobalSyncProducer(), ec, listenerMetrics)
}

func testBackOff() backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), maxRecoveryDefault)
}
