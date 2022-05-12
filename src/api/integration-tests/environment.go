package integrationtests

import (
	"context"
	"fmt"
	"github.com/consensys/orchestrate/pkg/sdk/messenger"
	"os"
	"path"
	"strconv"
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/notifier"
	"github.com/consensys/orchestrate/tests/pkg/trackers"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	"github.com/go-pg/pg/v9"

	"github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	webhook "github.com/consensys/orchestrate/src/infra/webhook/http"

	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	httputils "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api"
	"github.com/consensys/orchestrate/src/api/store/postgres/migrations"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/tests/pkg/docker"
	"github.com/consensys/orchestrate/tests/pkg/docker/config"
	ganacheDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/ganache"
	hashicorpDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/hashicorp"
	kafkaDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/kafka"
	postgresDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/postgres"
	quorumkeymanagerDocker "github.com/consensys/orchestrate/tests/pkg/docker/container/quorum-key-manager"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/zookeeper"
	integrationtest "github.com/consensys/orchestrate/tests/pkg/integration-test"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/traefik/traefik/v2/pkg/log"
	"gopkg.in/h2non/gock.v1"
)

const postgresContainerID = "postgres"
const qkmPostgresContainerID = "qkm-postgres"
const kafkaContainerID = "tx-sender-kafka-api"
const zookeeperContainerID = "tx-sender-zookeeper-api"
const ganacheContainerID = "tx-sender-ganache-api"
const qkmContainerID = "quorum-key-manager"
const qkmContainerMigrateID = "quorum-key-manager-migrate"
const hashicorpContainerID = "hashicorp"
const networkName = "api"
const qkmDefaultStoreID = "orchestrate-eth"
const hashicorpMountPath = "orchestrate"

// nolint
const waitForNotificationTimeOut = 5 * time.Second

var envPGHostPort string
var envQKMPGHostPort string
var envKafkaHostPort string
var envHTTPPort string
var envMetricsPort string
var envGanacheHostPort string
var envQKMHostPort string
var envVaultHostPort string

type IntegrationEnvironment struct {
	ctx                      context.Context
	logger                   log.Logger
	api                      *app.App
	client                   *docker.Client
	messengerConsumerTracker *trackers.MessengerConsumerTracker
	notifierConsumerTracker  *trackers.NotifierConsumerTracker
	notificationTopic        string
	baseURL                  string
	metricsURL               string
	apiCfg                   *api.Config
	notifierCfg              *notifier.Config
	blockchainNodeURL        string
}

func NewIntegrationEnvironment(ctx context.Context) (*IntegrationEnvironment, error) {
	logger := log.FromContext(ctx)
	envPGHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))
	envQKMPGHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))
	envHTTPPort = strconv.Itoa(utils.RandIntRange(20000, 28080))
	envMetricsPort = strconv.Itoa(utils.RandIntRange(30000, 38082))
	envKafkaHostPort = strconv.Itoa(utils.RandIntRange(20000, 29092))
	envGanacheHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))
	envQKMHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))
	envVaultHostPort = strconv.Itoa(utils.RandIntRange(10000, 15235))

	apiTopic := "topic-api"
	txSenderTopic := "topic-tx-sender"
	txListenerTopic := "topic-tx-listener"
	notifierTopic := "topic-notifier"
	decodedTopic := "topic-tx-decoded"

	// Define external hostname
	kafkaExternalHostname := fmt.Sprintf("localhost:%s", envKafkaHostPort)
	quorumKeyManagerURL := fmt.Sprintf("http://localhost:%s", envQKMHostPort)

	// Initialize environment flags
	flgs := pflag.NewFlagSet("api-integration-test", pflag.ContinueOnError)
	flags.NewAPIFlags(flgs)
	flags.NotifierFlags(flgs)
	args := []string{
		"--metrics-port=" + envMetricsPort,
		"--rest-port=" + envHTTPPort,
		"--db-port=" + envPGHostPort,
		"--kafka-url=" + kafkaExternalHostname,
		"--topic-api=" + apiTopic,
		"--topic-tx-listener=" + txListenerTopic,
		"--topic-tx-sender=" + txSenderTopic,
		"--topic-notifier=" + notifierTopic,
		"--key-manager-url=" + quorumKeyManagerURL,
		"--key-manager-store-name=" + qkmDefaultStoreID,
		"--log-level=panic",
	}

	err := flgs.Parse(args)
	if err != nil {
		logger.WithError(err).Error("cannot parse environment flags")
		return nil, err
	}

	rootToken := fmt.Sprintf("root_token_%v", strconv.Itoa(utils.RandIntRange(0, 10000)))
	vaultContainer := hashicorpDocker.NewDefault().
		SetHostPort(envVaultHostPort).
		SetRootToken(rootToken).
		SetMountPath(hashicorpMountPath)

	manifestsPath, err := getManifestsPath()
	if err != nil {
		return nil, err
	}

	qkmContainer := quorumkeymanagerDocker.NewDefault().
		SetHostPort(envQKMHostPort).
		SetDBHost(qkmPostgresContainerID).
		SetManifestDirectory(manifestsPath)

	qkmContainerMigrate := quorumkeymanagerDocker.NewDefaultMigrate().
		SetDBHost(qkmPostgresContainerID)

	err = qkmContainer.CreateManifest("manifest.yml", &http.Manifest{
		Kind:    "Ethereum",
		Version: "0.0.1",
		Name:    qkmDefaultStoreID,
		Specs: map[string]interface{}{
			"keystore": "HashicorpKeys",
			"specs": map[string]string{
				"mountPoint": hashicorpMountPath,
				"address":    "http://" + hashicorpContainerID + ":8200",
				"token":      rootToken,
				"namespace":  "",
			},
		},
	})

	if err != nil {
		return nil, err
	}

	// Initialize environment container setup
	composition := &config.Composition{
		Containers: map[string]*config.Container{
			postgresContainerID:    {Postgres: postgresDocker.NewDefault().SetHostPort(envPGHostPort)},
			qkmPostgresContainerID: {Postgres: postgresDocker.NewDefault().SetHostPort(envQKMPGHostPort)},
			zookeeperContainerID:   {Zookeeper: zookeeper.NewDefault()},
			kafkaContainerID: {Kafka: kafkaDocker.NewDefault().
				SetHostPort(envKafkaHostPort).
				SetZookeeperHostname(zookeeperContainerID).
				SetKafkaInternalHostname(kafkaContainerID).
				SetKafkaExternalHostname(kafkaExternalHostname),
			},
			hashicorpContainerID:  {HashicorpVault: vaultContainer},
			qkmContainerMigrateID: {QuorumKeyManagerMigrate: qkmContainerMigrate},
			qkmContainerID:        {QuorumKeyManager: qkmContainer},
			ganacheContainerID:    {Ganache: ganacheDocker.NewDefault().SetHostPort(envGanacheHostPort)},
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
		baseURL:           "http://localhost:" + envHTTPPort,
		metricsURL:        "http://localhost:" + envMetricsPort,
		apiCfg:            flags.NewAPIConfig(viper.GetViper()),
		notifierCfg:       flags.NewNotifierConfig(viper.GetViper()),
		blockchainNodeURL: fmt.Sprintf("http://localhost:%s", envGanacheHostPort),
		notificationTopic: decodedTopic,
	}, nil
}

func (env *IntegrationEnvironment) Start(ctx context.Context) error {
	err := env.client.CreateNetwork(ctx, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not create network")
		return err
	}

	// Start Hashicorp Vault
	err = env.client.Up(ctx, hashicorpContainerID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up vault container")
		return err
	}

	err = env.client.WaitTillIsReady(ctx, hashicorpContainerID, 10*time.Second)
	if err != nil {
		env.logger.WithError(err).Error("could not start vault")
		return err
	}

	err = env.client.Up(ctx, qkmPostgresContainerID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up QKM postgres")
		return err
	}

	// Start postgres Database
	err = env.client.Up(ctx, postgresContainerID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up postgres")
		return err
	}

	err = env.client.WaitTillIsReady(ctx, postgresContainerID, 10*time.Second)
	if err != nil {
		env.logger.WithError(err).Error("could not start postgres")
		return err
	}

	// Start quorum key manager migration
	err = env.client.Up(ctx, qkmContainerMigrateID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up quorum-key-manager")
		return err
	}

	// Start quorum key manager
	err = env.client.Up(ctx, qkmContainerID, networkName)
	if err != nil {
		env.logger.WithError(err).Error("could not up quorum-key-manager")
		return err
	}

	err = env.client.WaitTillIsReady(ctx, qkmContainerID, 10*time.Second)
	if err != nil {
		env.logger.WithError(err).Error("could not start quorum-key-manager")
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

	// Run postgres migrations
	err = env.migrate()
	if err != nil {
		env.logger.WithError(err).Error("could not migrate postgres")
		return err
	}

	env.api, err = newAPI(ctx, env.apiCfg, env.notifierCfg)
	if err != nil {
		env.logger.WithError(err).Error("could initialize API")
		return err
	}

	// Start internal kafka consumer
	env.messengerConsumerTracker, err = trackers.NewMessengerConsumerTracker(*env.apiCfg.Kafka, []string{env.apiCfg.KafkaTopics.Sender, env.apiCfg.KafkaTopics.Listener, env.apiCfg.KafkaTopics.Notifier})
	if err != nil {
		env.logger.WithError(err).Error("could initialize kafka internal Consumer")
		return err
	}

	// Start internal kafka consumer
	env.notifierConsumerTracker, err = trackers.NewNotifierConsumerTacker(*env.apiCfg.Kafka, []string{env.notificationTopic})
	if err != nil {
		env.logger.WithError(err).Error("could initialize kafka internal Consumer")
		return err
	}

	go func() {
		_ = env.messengerConsumerTracker.Consume(ctx)
	}()

	go func() {
		_ = env.notifierConsumerTracker.Consume(ctx)
	}()

	// Start API
	err = env.api.Start(ctx)
	if err != nil {
		env.logger.WithError(err).Error("could not start API")
		return err
	}

	integrationtest.WaitForServiceLive(
		ctx,
		fmt.Sprintf("%s/live", env.metricsURL),
		"api",
		15*time.Second,
	)

	// Wait for consumers to be ready
	time.Sleep(time.Second * 10)

	return nil
}

func (env *IntegrationEnvironment) Teardown(ctx context.Context) {
	env.logger.Info("tearing test suite down")

	if env.api != nil {
		err := env.api.Stop(ctx)
		if err != nil {
			env.logger.WithError(err).Error("could not stop API")
		}
	}

	err := env.client.Down(ctx, qkmContainerMigrateID)
	if err != nil {
		env.logger.WithError(err).Error("could not down quorum-key-manager-migration")
	}

	err = env.client.Down(ctx, qkmContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down quorum-key-manager")
	}

	err = env.client.Down(ctx, hashicorpContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down zookeeper")
	}

	err = env.client.Down(ctx, ganacheContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down ganache")
	}

	err = env.client.Down(ctx, qkmPostgresContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down qkm postgres")
	}

	err = env.client.Down(ctx, postgresContainerID)
	if err != nil {
		env.logger.WithError(err).Error("could not down postgres")
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

func (env *IntegrationEnvironment) migrate() error {
	pgCfg, err := flags.NewPGConfig(viper.GetViper()).ToPGOptionsV9()
	if err != nil {
		return err
	}

	pgDB := pg.Connect(pgCfg)

	_, _, err = migrations.Run(pgDB, "init")
	if err != nil {
		return err
	}

	_, _, err = migrations.Run(pgDB, "up")
	if err != nil {
		return err
	}

	err = pgDB.Close()
	if err != nil {
		return err
	}

	return nil
}

func newAPI(ctx context.Context, cfg *api.Config, notifierConfig *notifier.Config) (*app.App, error) {
	qkmClient, err := http.New(cfg.QKM)
	if err != nil {
		return nil, err
	}

	postgresClient, err := gopg.New("orchestrate.integration-tests.api", cfg.Postgres)
	if err != nil {
		return nil, err
	}

	kafkaProducer, err := sarama.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, err
	}

	messengerClient := messenger.NewProducerClient(&messenger.Config{
		TopicAPI:        notifierConfig.TopicAPI,
		TopicTxListener: cfg.KafkaTopics.Listener,
		TopicTxSender:   cfg.KafkaTopics.Sender,
		TopicTxNotifier: cfg.KafkaTopics.Notifier,
	}, kafkaProducer)

	authjwt.Init(ctx)
	authkey.Init(ctx)
	ethclient.Init(ctx)
	orchestrateclient.Init()

	interceptedHTTPClient := httputils.NewClient(httputils.NewDefaultConfig())
	gock.InterceptClient(interceptedHTTPClient)
	webhookProducer := webhook.New(interceptedHTTPClient)

	notifierDaemon, err := notifier.New(
		notifierConfig,
		kafkaProducer,
		webhookProducer,
		messengerClient,
	)
	if err != nil {
		return nil, err
	}

	return api.NewAPI(
		cfg,
		postgresClient,
		authjwt.GlobalChecker(),
		authkey.GlobalChecker(),
		qkmClient,
		cfg.QKM.StoreName,
		ethclient.GlobalClient(),
		messengerClient,
		notifierDaemon,
	)
}

func getManifestsPath() (string, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(currDir, "manifests"), nil
}
