package api

import (
	"context"

	broker "github.com/consensys/orchestrate/src/infra/kafka/sarama"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	qkmhttp "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	nonclient "github.com/consensys/orchestrate/src/infra/quorum-key-manager/non-client"
	"github.com/consensys/quorum-key-manager/pkg/client"
)

type Daemon struct {
	*app.App
}

// New Utility function used to initialize a new service
func New(ctx context.Context, cfg *Config) (*Daemon, error) {
	logger := log.FromContext(ctx)

	// Initialize infra dependencies
	qkmClient, err := QKMClient(cfg)
	if err != nil {
		return nil, err
	}

	postgresClient, err := gopg.New("orchestrate.api", cfg.Postgres)
	if err != nil {
		return nil, err
	}

	authjwt.Init(ctx)
	authkey.Init(ctx)
	ethclient.Init(ctx)

	clientProducer, err := broker.NewProducer(cfg.Kafka)
	if err != nil {
		return nil, err
	}
	logger.WithField("host", cfg.Kafka.URLs).Info("kafka producer client ready")

	api, err := NewAPI(
		cfg,
		postgresClient,
		authjwt.GlobalChecker(),
		authkey.GlobalChecker(),
		qkmClient,
		cfg.QKM.StoreName,
		ethclient.GlobalClient(),
		clientProducer,
		cfg.KafkaTopics,
	)

	if err != nil {
		return nil, err
	}

	return &Daemon{
		api,
	}, nil
}

func (d *Daemon) Run(ctx context.Context) error {
	return d.App.Run(ctx)
}

func QKMClient(cfg *Config) (client.KeyManagerClient, error) {
	if cfg.QKM.URL != "" {
		return qkmhttp.New(cfg.QKM)
	}

	return nonclient.NewNonClient(), nil
}
