package txlistener

import (
	"time"

	"github.com/consensys/orchestrate/src/infra/messenger/kafka"

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
	Kafka                 *kafka.Config
}
