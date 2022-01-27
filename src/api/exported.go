package api

import (
	"context"

	authjwt "github.com/consensys/orchestrate/pkg/toolkit/app/auth/jwt"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"

	qkm "github.com/consensys/orchestrate/src/infra/quorum-key-manager"

	"github.com/consensys/orchestrate/pkg/toolkit/app"
	authkey "github.com/consensys/orchestrate/pkg/toolkit/app/auth/key"
	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/spf13/viper"
)

// New Utility function used to initialize a new service
func New(ctx context.Context) (*app.App, error) {
	// Initialize dependencies
	authjwt.Init(ctx)
	authkey.Init(ctx)
	sarama.InitSyncProducer(ctx)
	ethclient.Init(ctx)
	qkm.Init()

	config := NewConfig(viper.GetViper())
	pgmngr := postgres.GetManager()

	return NewAPI(
		config,
		pgmngr,
		authjwt.GlobalChecker(),
		authkey.GlobalChecker(),
		qkm.GlobalClient(),
		qkm.GlobalStoreName(),
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
