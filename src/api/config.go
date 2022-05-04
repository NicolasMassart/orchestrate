package api

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/src/api/proxy"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
)

type TopicConfig struct {
	API      string
	Sender   string
	Listener string
	Notifier string
}

type Config struct {
	App          *app.Config
	Postgres     *gopg.Config
	Multitenancy bool
	Proxy        *proxy.Config
	QKM          *quorumkeymanager.Config
	Kafka        *kafka.Config
	KafkaTopics  *TopicConfig
}
