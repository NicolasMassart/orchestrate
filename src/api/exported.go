package api

import (
	"context"
	"net/http"

	"github.com/consensys/orchestrate/pkg/sdk/messenger"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	qkmhttp "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	nonclient "github.com/consensys/orchestrate/src/infra/quorum-key-manager/non-client"
	webhook "github.com/consensys/orchestrate/src/infra/webhook/http"
	"github.com/consensys/orchestrate/src/notifier"
	qkmclient "github.com/consensys/quorum-key-manager/pkg/client"
)

type Daemon struct {
	*app.App
}

func New(ctx context.Context, cfg *Config, notifierCfg *notifier.Config) (*Daemon, error) {
	// Initialize infra dependencies
	qkmClient, err := QKMClient(cfg)
	if err != nil {
		return nil, err
	}

	postgresClient, err := gopg.New("orchestrate.api", cfg.Postgres)
	if err != nil {
		return nil, err
	}

	kafkaProdClient, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, err
	}

	// @TODO Decouple initialization of api and notifier to prevent this overhead of merging topics
	messengerClient := messenger.NewProducerClient(&messenger.Config{
		TopicAPI:        notifierCfg.Messenger.TopicAPI,
		TopicTxListener: cfg.Messenger.TopicTxListener,
		TopicTxSender:   cfg.Messenger.TopicTxSender,
		TopicNotifier:   notifierCfg.ConsumerTopic,
	}, kafkaProdClient)
	webhookProducer := webhook.New(http.DefaultClient)

	authjwt.Init(ctx)
	authkey.Init(ctx)
	ethclient.Init(ctx)
	client.Init()

	notifierDaemon, err := notifier.New(notifierCfg, kafkaProdClient, webhookProducer, messengerClient)
	if err != nil {
		return nil, err
	}

	api, err := NewAPI(
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

	if err != nil {
		return nil, err
	}

	return &Daemon{api}, nil
}

func (d *Daemon) Run(ctx context.Context) error {
	return d.App.Run(ctx)
}

func QKMClient(cfg *Config) (qkmclient.KeyManagerClient, error) {
	if cfg.QKM.URL != "" {
		return qkmhttp.New(cfg.QKM)
	}

	return nonclient.NewNonClient(), nil
}
