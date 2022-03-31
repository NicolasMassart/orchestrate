package integrationtests

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	httputils "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/kafka"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/redis"
	"github.com/consensys/orchestrate/src/infra/redis/redigo"
	txsender "github.com/consensys/orchestrate/src/tx-sender"
	"github.com/consensys/orchestrate/src/tx-sender/store"
	noncesender "github.com/consensys/orchestrate/src/tx-sender/store/redis"
	"github.com/consensys/orchestrate/tests/pkg/docker"
	"github.com/consensys/orchestrate/tests/pkg/docker/config"
	kafkaDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/kafka"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/zookeeper"
	integrationtest "github.com/consensys/orchestrate/tests/pkg/integration-test"
	qkmclient "github.com/consensys/quorum-key-manager/pkg/client"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/traefik/traefik/v2/pkg/log"
	"gopkg.in/h2non/gock.v1"
)

const kafkaContainerID = "Kafka-tx-sender"
const zookeeperContainerID = "zookeeper-tx-sender"
const apiURL = "http://api:8081"
const keyManagerURL = "http://key-manager:8081"
const apiMetricsURL = "http://api:8082"
const networkName = "tx-sender"
const qkmStoreName = "orchestrate-eth"
const maxRecoveryDefault = 1

var envKafkaHostPort string
var envMetricsPort string

type IntegrationEnvironment struct {
	ctx         context.Context
	logger      log.Logger
	txSender    *app.App
	client      *docker.Client
	producer    kafka.Producer
	metricsURL  string
	ns          store.NonceSender
	redis       redis.Client
	txSenderCfg *txsender.Config
}

func NewIntegrationEnvironment(ctx context.Context) (*IntegrationEnvironment, error) {
	logger := log.FromContext(ctx)
	envMetricsPort = strconv.Itoa(utils.RandIntRange(30000, 38082))
	envKafkaHostPort = strconv.Itoa(utils.RandIntRange(20000, 29092))

	// Define external hostname
	kafkaExternalHostname := os.Getenv("KAFKA_HOST")
	if kafkaExternalHostname == "" {
		kafkaExternalHostname = "localhost"
	}

	kafkaExternalHostname = fmt.Sprintf("%s:%s", kafkaExternalHostname, envKafkaHostPort)

	// Initialize environment flags
	flgs := pflag.NewFlagSet("tx-sender-integration-test", pflag.ContinueOnError)
	flags.TxSenderFlags(flgs)
	args := []string{
		"--metrics-port=" + envMetricsPort,
		"--kafka-url=" + kafkaExternalHostname,
		"--nonce-manager-type=" + txsender.NonceManagerTypeRedis,
		"--key-manager-url=" + keyManagerURL,
		"--key-manager-store-name=" + qkmStoreName,
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
		},
	}

	dockerClient, err := docker.NewClient(composition)
	if err != nil {
		logger.WithError(err).Error("cannot initialize docker client")
		return nil, err
	}

	mredis, err := miniredis.Run()
	if err != nil {
		logger.WithError(err).Error("cannot initialize miniredis")
		return nil, err
	}

	redisClient, err := redigo.New(&redigo.Config{Host: mredis.Host(), Port: mredis.Port()})
	if err != nil {
		logger.WithError(err).Error("cannot initialize redis client")
		return nil, err
	}

	return &IntegrationEnvironment{
		ctx:         ctx,
		logger:      logger,
		client:      dockerClient,
		metricsURL:  "http://localhost:" + envMetricsPort,
		txSenderCfg: flags.NewTxSenderConfig(viper.GetViper()),
		redis:       redisClient,
		ns:          noncesender.NewNonceSender(redisClient, 100000),
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

	consumer, err := sarama.NewConsumer(env.txSenderCfg.Kafka)
	if err != nil {
		return err
	}

	// Create app
	env.txSenderCfg.BckOff = testBackOff()
	env.txSender, err = newTxSender(env.txSenderCfg, env.redis, consumer)
	if err != nil {
		env.logger.WithError(err).Error("could not initialize tx-sender")
		return err
	}

	env.producer, err = sarama.NewProducer(env.txSenderCfg.Kafka)
	if err != nil {
		env.logger.WithError(err).Error("could not initialize kafka producer")
		return err
	}

	// Start tx-sender app
	err = env.txSender.Start(ctx)
	if err != nil {
		env.logger.WithError(err).Error("could not start tx-sender")
		return err
	}

	env.logger.Debug("Waiting for consumer to start....")
	time.Sleep(time.Second * 5) // @TODO Improve awaiting trigger

	integrationtest.WaitForServiceLive(ctx, fmt.Sprintf("%s/live", env.metricsURL), "tx-sender", 15*time.Second)
	return nil
}

func (env *IntegrationEnvironment) Teardown(ctx context.Context) {
	env.logger.Info("tearing test suite down")

	err := env.txSender.Stop(ctx)
	if err != nil {
		env.logger.WithError(err).Error("could not stop tx-sender")
	}

	err = env.client.Down(ctx, kafkaContainerID)
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

func newTxSender(txSenderConfig *txsender.Config, redisCli redis.Client, consumer kafka.Consumer) (*app.App, error) {
	// Initialize dependencies
	httpClient := httputils.NewClient(httputils.NewDefaultConfig())
	gock.InterceptClient(httpClient)

	ec := ethclient.NewClient(testBackOff, httpClient)
	qkmClient := qkmclient.NewHTTPClient(httpClient, &qkmclient.Config{
		URL: keyManagerURL,
	})

	conf2 := client.NewConfig(apiURL, "", nil)
	conf2.MetricsURL = apiMetricsURL
	apiClient := client.NewHTTPClient(httpClient, conf2)

	txSenderConfig.NonceMaxRecovery = maxRecoveryDefault

	return txsender.NewTxSender(txSenderConfig,
		[]kafka.Consumer{consumer},
		qkmClient,
		apiClient,
		ec,
		redisCli)
}

func testBackOff() backoff.BackOff {
	return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), maxRecoveryDefault)
}
