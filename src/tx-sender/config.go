package txsender

import (
	"time"

	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"

	"github.com/consensys/orchestrate/src/infra/redis/redigo"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

const (
	NonceManagerTypeInMemory = "in-memory"
	NonceManagerTypeRedis    = "redis"
)

type Config struct {
	App                    *app.Config
	Kafka                  *sarama.Config
	KafkaTopicTxSender     string
	GroupName              string
	NConsumer              int
	ProxyURL               string
	BckOff                 backoff.BackOff
	NonceMaxRecovery       uint64
	NonceManagerType       string
	IsMultiTenancyEnabled  bool
	RedisCfg               *redigo.Config
	NonceManagerExpiration time.Duration
	QKM                    *quorumkeymanager.Config
}
