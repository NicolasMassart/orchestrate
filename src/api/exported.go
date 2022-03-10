package api

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	qkmhttp "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	nonclient "github.com/consensys/orchestrate/src/infra/quorum-key-manager/non-client"
	"github.com/consensys/quorum-key-manager/pkg/client"
	"github.com/spf13/viper"
)

// New Utility function used to initialize a new service
func New(ctx context.Context) (*app.App, error) {
	cfg := NewConfig(viper.GetViper())

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
	sarama.InitSyncProducer(ctx)
	ethclient.Init(ctx)

	return NewAPI(
		cfg,
		postgresClient,
		authjwt.GlobalChecker(),
		authkey.GlobalChecker(),
		qkmClient,
		cfg.QKM.StoreName,
		ethclient.GlobalClient(),
		sarama.GlobalSyncProducer(),
		sarama.NewKafkaTopicConfig(viper.GetViper()),
	)
}

func Run(ctx context.Context) error {
	appli, err := New(ctx)
	if err != nil {
		return err
	}
	return appli.Run(ctx)
}

func QKMClient(cfg *Config) (client.KeyManagerClient, error) {
	if cfg.QKM.URL != "" {
		return qkmhttp.New(cfg.QKM)
	}

	return nonclient.NewNonClient(), nil
}
