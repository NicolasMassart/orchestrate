package txlistener

import (
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/messenger"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
)

type Config struct {
	IsMultiTenancyEnabled bool
	App                   *app.Config
	HTTPClient            *http.Config
	API                   *orchestrateclient.Config
	RetryInterval         time.Duration
	ConsumerTopic         string
	Messenger             *messenger.Config
	Kafka                 *kafka.Config
}
