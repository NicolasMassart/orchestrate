package api

import (
	"github.com/consensys/orchestrate/pkg/sdk/messenger"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/src/api/proxy"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/postgres/gopg"
	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
)

type Config struct {
	App          *app.Config
	Postgres     *gopg.Config
	Multitenancy bool
	Proxy        *proxy.Config
	QKM          *quorumkeymanager.Config
	Kafka        *kafka.Config
	Messenger    *messenger.Config
}
