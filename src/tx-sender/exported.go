package txsender

import (
	"context"

	broker "github.com/consensys/orchestrate/src/infra/kafka"
	brokerClient "github.com/consensys/orchestrate/src/infra/kafka/sarama"

	qkmhttp "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	nonclient "github.com/consensys/orchestrate/src/infra/quorum-key-manager/non-client"
	"github.com/consensys/orchestrate/src/infra/redis"
	"github.com/consensys/orchestrate/src/infra/redis/redigo"
	"github.com/consensys/quorum-key-manager/pkg/client"

	orchestrateClient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
)

// New Utility function used to initialize a new service
func New(ctx context.Context, cfg *Config) (*app.App, error) {
	logger := log.FromContext(ctx)

	// Initialize infra dependencies
	redisClient, err := getRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	qkmClient, err := getQKMClient(cfg)
	if err != nil {
		return nil, err
	}

	orchestrateClient.Init()
	ethclient.Init(ctx)

	consumerGroups := make([]broker.Consumer, cfg.NConsumer)
	for idx := 0; idx < cfg.NConsumer; idx++ {
		consumerGroups[idx], err = brokerClient.NewConsumer(cfg.Kafka)
		if err != nil {
			return nil, err
		}

		logger.WithField("host", cfg.Kafka.URLs).WithField("group_name", cfg.GroupName).Info("consumer client ready")
	}

	return NewTxSender(
		cfg,
		consumerGroups,
		qkmClient,
		orchestrateClient.GlobalClient(),
		ethclient.GlobalClient(),
		redisClient,
	)
}

func getRedisClient(cfg *Config) (redis.Client, error) {
	if cfg.NonceManagerType == NonceManagerTypeRedis {
		return redigo.New(cfg.RedisCfg)
	}

	return nil, nil
}

func getQKMClient(cfg *Config) (client.KeyManagerClient, error) {
	if cfg.QKM.URL != "" {
		return qkmhttp.New(cfg.QKM)
	}

	return nonclient.NewNonClient(), nil
}
