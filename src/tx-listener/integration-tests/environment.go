package integrationtests

import (
	"context"
	"fmt"
	"math/big"
	http2 "net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/sdk/messenger"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	httputils "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	ethclient2 "github.com/consensys/orchestrate/src/infra/ethclient"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	txlistener "github.com/consensys/orchestrate/src/tx-listener"
	listenermetrics "github.com/consensys/orchestrate/src/tx-listener/tx-listener/metrics"
	"github.com/consensys/orchestrate/tests/pkg/docker"
	"github.com/consensys/orchestrate/tests/pkg/docker/config"
	ganacheDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/ganache"
	kafkaDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/kafka"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/zookeeper"
	"github.com/consensys/orchestrate/tests/pkg/trackers"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

const apiURL = "http://api:8081"
const apiMetricsURL = "http://api:8082"
const networkName = "tx-listener"
const maxRecoveryDefault = 1
const ganacheContainerID = "tx-listener-ganache-api"
const kafkaContainerID = "Kafka-tx-listener"
const zookeeperContainerID = "zookeeper-tx-listener"
const ganacheChainUUID = "ganacheChainUUID"
const ganacheChainID = "666"

var envMetricsPort string
var envGanacheHostPort string
var envKafkaHostPort string

type IntegrationEnvironment struct {
	ctx                      context.Context
	cancel                   context.CancelFunc
	T                        *testing.T
	logger                   *log.Logger
	app                      *app.App
	messengerClient          sdk.MessengerTxListener
	messengerConsumerTracker *trackers.MessengerConsumerTracker
	ethClient                ethclient2.MultiClient
	client                   *docker.Client
	cfg                      *txlistener.Config
	chain                    *entities.Chain
	blockchainNodeURL        string
	proxyURL                 string
	err                      error
}

func NewIntegrationEnvironment(ctx context.Context, cancel context.CancelFunc, t *testing.T) (*IntegrationEnvironment, error) {
	logger := log.NewLogger()
	envMetricsPort = strconv.Itoa(utils.RandIntRange(30000, 38082))
	envGanacheHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))
	envKafkaHostPort = strconv.Itoa(utils.RandIntRange(20000, 29092))

	// Define external hostname
	kafkaExternalHostname := "localhost"
	kafkaExternalHostname = fmt.Sprintf("%s:%s", kafkaExternalHostname, envKafkaHostPort)

	// Initialize environment flags
	flgs := pflag.NewFlagSet("tx-listener-integration-test", pflag.ContinueOnError)
	flags.TxListenerFlags(flgs)
	args := []string{
		"--metrics-port=" + envMetricsPort,
		"--kafka-url=" + kafkaExternalHostname,
		"--api-url=" + apiURL,
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
			ganacheContainerID:   {Ganache: ganacheDocker.NewDefault().SetHostPort(envGanacheHostPort).SetChainID(ganacheChainID)},
			zookeeperContainerID: {Zookeeper: zookeeper.NewDefault()},
			kafkaContainerID: {Kafka: kafkaDocker.NewDefault().
				SetHostPort(envKafkaHostPort).
				SetZookeeperHostname(zookeeperContainerID).
				SetKafkaInternalHostname(kafkaContainerID).
				SetKafkaExternalHostname(kafkaExternalHostname),
			},
		},
	}

	dockerClient, err := docker.NewClient(composition)
	if err != nil {
		logger.WithError(err).Error("cannot initialize docker client")
		return nil, err
	}

	return &IntegrationEnvironment{
		T:                 t,
		ctx:               ctx,
		cancel:            cancel,
		logger:            logger,
		client:            dockerClient,
		cfg:               flags.NewTxListenerConfig(viper.GetViper()),
		blockchainNodeURL: fmt.Sprintf("http://localhost:%s", envGanacheHostPort),
		proxyURL:          fmt.Sprintf("%s/proxy/chains/%s", apiURL, ganacheChainUUID),
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

	apiClient := newAPIClient()

	env.ethClient, err = newEthClient(env.blockchainNodeURL)
	if err != nil {
		env.logger.WithError(err).Error("could not initialize eth client")
		return err
	}

	kafkaProd, err := kafka.NewProducer(env.cfg.Kafka)
	if err != nil {
		env.logger.WithError(err).Error("could not initialize kafka producer")
		return err
	}
	
	// Start internal kafka consumer
	env.messengerConsumerTracker, err = trackers.NewMessengerConsumerTracker(*env.cfg.Kafka, []string{env.cfg.Messenger.TopicAPI})
	if err != nil {
		env.logger.WithError(err).Error("could initialize kafka internal Consumer")
		return err
	}
	
	go func() {
		_ = env.messengerConsumerTracker.Consume(ctx)
	}()

	// Create app
	env.app, err = txlistener.NewTxListener(env.cfg, kafkaProd, apiClient, env.ethClient, listenermetrics.NewListenerNopMetrics())

	env.chain = newChain(env.blockchainNodeURL)

	env.messengerClient = messenger.NewProducerClient(&messenger.Config{TopicTxListener: env.cfg.ConsumerTopic}, kafkaProd)

	// Start tx-sender app
	err = env.app.Start(ctx)
	if err != nil {
		env.logger.WithError(err).Error("could not start tx-sender")
		return err
	}

	env.logger.Debug("Waiting for consumer to start....")
	time.Sleep(time.Second * 5) // @TODO Improve awaiting trigger

	return nil
}

func (env *IntegrationEnvironment) Teardown(ctx context.Context) {
	env.logger.Info("tearing test suite down")
	env.cancel()

	err := env.app.Stop(ctx)
	if err != nil {
		env.logger.WithError(err).Error("could delete chain")
	}

	err = env.client.Down(ctx, kafkaContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down Kafka")
	}

	err = env.client.Down(ctx, zookeeperContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down zookeeper")
	}

	err = env.client.Down(ctx, ganacheContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down ganache")
	}

	err = env.client.RemoveNetwork(ctx, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not remove network")
	}

	require.NoError(env.T, err)
}

func newAPIClient() *client.HTTPClient {
	httpClient := httputils.NewClient(httputils.NewDefaultConfig())
	gock.InterceptClient(httpClient)

	conf2 := client.NewConfig(apiURL, "", nil)
	conf2.MetricsURL = apiMetricsURL
	return client.NewHTTPClient(httpClient, conf2)
}

func newChain(blockchainURL string) *entities.Chain {
	chain := testdata.FakeChain()
	chain.UUID = ganacheChainUUID
	chain.ChainID, _ = new(big.Int).SetString(ganacheChainID, 10)
	chain.ListenerBlockTimeDuration = time.Second
	chain.URLs = []string{blockchainURL}
	return chain
}

func newEthClient(blockchainURL string) (ethclient2.MultiClient, error) {
	proxyURL, err := url.Parse(blockchainURL)
	if err != nil {
		return nil, err
	}

	ec := ethclient.NewClientWithBackOff(func() backoff.BackOff {
		return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), maxRecoveryDefault)
	}, &http2.Client{
		Transport: &http2.Transport{
			Proxy: http2.ProxyURL(proxyURL),
		},
	})
	return ec, nil
}
