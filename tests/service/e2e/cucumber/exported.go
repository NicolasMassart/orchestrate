package cucumber

import (
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/kafka/testutils"
	"github.com/consensys/orchestrate/tests/service/e2e/cucumber/steps"
	"github.com/cucumber/godog"
	log "github.com/sirupsen/logrus"
)

func Run(opt *godog.Options, consumerTracker *testutils.ConsumerTracker, topics *sarama.TopicConfig, waitFor time.Duration) error {
	status := godog.TestSuite{
		Name: "tests",
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			steps.InitializeScenario(s, consumerTracker, topics, waitFor)
		},
		Options: opt,
	}.Run()

	// godog status:
	//  0 - success
	//  1 - failed
	//  2 - command line usage error
	//  128 - or higher, os signal related error exit codes

	// If fail
	if status > 0 {
		return errors.InternalError("cucumber: feature tests failed with status %d", status)
	}

	log.Info("cucumber: feature tests succeeded")
	return nil
}
