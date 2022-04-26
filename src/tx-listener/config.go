package txlistener

import (
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

type Config struct {
	IsMultiTenancyEnabled bool
	App                   *app.Config
	HTTPClient            *http.Config
	API                   *orchestrateclient.Config
	RetryInterval         time.Duration
	ConsumerTopic         string
	Kafka                 *sarama.Config
}
