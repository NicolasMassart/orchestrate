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
	"github.com/consensys/orchestrate/src/api/service/formatters"
	chainlistener "github.com/consensys/orchestrate/src/chain-listener"
	chain_listener "github.com/consensys/orchestrate/src/chain-listener/chain-listener"
	"github.com/consensys/orchestrate/src/chain-listener/chain-listener/builder"
	"github.com/consensys/orchestrate/src/chain-listener/service/listener"
	datapullers "github.com/consensys/orchestrate/src/chain-listener/service/listener/data-pullers"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	ethclient2 "github.com/consensys/orchestrate/src/infra/ethclient"
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
const ganacheChainUUID = "ganacheChainUUID"
const ganacheChainID = "666"

var envKafkaHostPort string
var envMetricsPort string
var envGanacheHostPort string

type IntegrationEnvironment struct {
	ctx                context.Context
	cancel             context.CancelFunc
	T                  *testing.T
	logger             *log.Logger
	ucs                chain_listener.EventUseCases
	chainBlockListener listener.ChainBlockListener
	ethClient          ethclient2.MultiClient
	client             *docker.Client
	consumer           *integrationtest.KafkaConsumer
	producer           sarama2.SyncProducer
	cfg                *chainlistener.Config
	chain              *entities.Chain
	blockchainNodeURL  string
	proxyURL           string
}

func NewIntegrationEnvironment(ctx context.Context, cancel context.CancelFunc, t *testing.T) (*IntegrationEnvironment, error) {
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
			ganacheContainerID: {Ganache: ganacheDocker.NewDefault().SetHostPort(envGanacheHostPort).SetChainID(ganacheChainID)},
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
		cfg:               flags.NewChainListenerConfig(viper.GetViper()),
		producer:          sarama.GlobalSyncProducer(),
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

	err = env.client.WaitTillIsReady(ctx, kafkaContainerID, 20*time.Second)
	if err != nil {
		env.logger.WithError(err).Error("could not start Kafka")
		return err
	}

	sarama.InitSyncProducer(ctx)
	err = sarama.InitClient(ctx)
	if err != nil {
		env.logger.WithError(err).Error("cannot initialize kafka client")
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

	env.ethClient, err = newEthClient(env.blockchainNodeURL)
	if err != nil {
		env.logger.WithError(err).Error("could not initialize eth client")
		return err
	}

	// Start Kafka consumer
	env.consumer, err = integrationtest.NewKafkaTestConsumer(
		ctx,
		"chain-listener-integration-listener-group",
		sarama.GlobalClient(),
		[]string{env.cfg.ChainListenerConfig.DecodedOutTopic},
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

	apiClient := newAPIClient()
	// Create app
	env.ucs = newEventUseCases(ctx, env.cfg, apiClient, env.ethClient, env.logger)

	env.chain = newChain(env.blockchainNodeURL)

	env.chainBlockListener = datapullers.NewListenBlockSession(apiClient, env.ethClient, env.ucs.ChainBlockTxsUseCase(),
		env.chain, env.logger)

	err = env.ucs.AddChainUseCase().Execute(ctx, env.chain)
	if err != nil {
		env.logger.WithError(err).Error("could not add chain")
		return err
	}

	go func() {
		gock.New(apiURL).
			Patch("/chains/" + ganacheChainUUID).
			Times(-1).
			Reply(http2.StatusOK).JSON(formatters.FormatChainResponse(env.chain))

		_ = env.chainBlockListener.Run(ctx)
	}()

	return nil
}

func (env *IntegrationEnvironment) Teardown(ctx context.Context) {
	env.logger.Info("tearing test suite down")
	env.cancel()

	err := env.ucs.DeleteChainUseCase().Execute(ctx, env.chain.UUID)
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
	chain.ListenerCurrentBlock = 0
	chain.ListenerBackOffDuration = time.Second
	chain.URLs = []string{blockchainURL}
	return chain
}

func newEventUseCases(ctx context.Context, cfg *chainlistener.Config, apiClient client.OrchestrateClient,
	ec ethclient2.MultiClient, logger *log.Logger) chain_listener.EventUseCases {
	// Initialize dependencies
	sarama.InitSyncProducer(ctx)
	sarama.InitConsumerGroup(ctx, cfg.ChainListenerConfig.DecodedOutTopic)

	return builder.NewEventUseCases(apiClient, sarama.GlobalSyncProducer(), ec,
		cfg.ChainListenerConfig.DecodedOutTopic, logger)
}

func newEthClient(blockchainURL string) (ethclient2.MultiClient, error) {
	proxyURL, err := url.Parse(blockchainURL)
	if err != nil {
		return nil, err
	}

	ec := ethclient.NewClient(func() backoff.BackOff {
		return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), maxRecoveryDefault)
	}, &http2.Client{
		Transport: &http2.Transport{
			Proxy: http2.ProxyURL(proxyURL),
		},
	})
	return ec, nil
}