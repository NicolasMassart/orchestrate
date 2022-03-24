package txsender

import (
	"context"

	qkmhttp "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	nonclient "github.com/consensys/orchestrate/src/infra/quorum-key-manager/non-client"
	"github.com/consensys/orchestrate/src/infra/redis"
	"github.com/consensys/orchestrate/src/infra/redis/redigo"
	"github.com/consensys/quorum-key-manager/pkg/client"

	sarama2 "github.com/Shopify/sarama"
	orchestrateClient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"

	"github.com/consensys/orchestrate/src/infra/broker/sarama"
	"github.com/spf13/viper"
)

// New Utility function used to initialize a new service
func New(ctx context.Context) (*app.App, error) {
	logger := log.FromContext(ctx)
	cfg := NewConfig(viper.GetViper())

	// Initialize infra dependencies
	redisClient, err := getRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	qkmClient, err := getQKMClient(cfg)
	if err != nil {
		return nil, err
	}

	sarama.InitSyncProducer(ctx)
	orchestrateClient.Init()
	ethclient.Init(ctx)

	consumerGroups := make([]sarama2.ConsumerGroup, cfg.NConsumer)
	hostnames := viper.GetStringSlice(sarama.KafkaURLViperKey)
	for idx := 0; idx < cfg.NConsumer; idx++ {
		consumerGroups[idx], err = NewSaramaConsumer(hostnames, cfg.GroupName)
		if err != nil {
			return nil, err
		}
		logger.WithField("host", hostnames).WithField("group_name", cfg.GroupName).
			Info("consumer client ready")
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

func NewSaramaConsumer(hostnames []string, groupName string) (sarama2.ConsumerGroup, error) {
	cfg, err := sarama.NewSaramaConfig()
	if err != nil {
		return nil, err
	}

	saramaClient, err := sarama.NewClient(hostnames, cfg)
	if err != nil {
		return nil, err
	}

	return sarama.NewConsumerGroupFromClient(groupName, saramaClient)
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
