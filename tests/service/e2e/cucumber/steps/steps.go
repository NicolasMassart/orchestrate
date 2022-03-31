package steps

import (
	gohttp "net/http"
	"sync"
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	rpcClient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/kafka/proto"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/src/infra/kafka/testutils"
	"github.com/consensys/orchestrate/tests/service/e2e/cucumber/alias"
	"github.com/cucumber/godog"
	gherkin "github.com/cucumber/messages-go/v10"
	"github.com/mitchellh/copystructure"
)

// ScenarioContext is container for scenario context data
type ScenarioContext struct {
	Pickle *gherkin.Pickle

	consumerTracker *testutils.ConsumerTracker

	httpClient   *gohttp.Client
	httpResponse *gohttp.Response

	aliases *alias.Registry

	topics *sarama.TopicConfig

	waitForEnvelope time.Duration

	txResponses []*proto.TxResponse

	// API
	client orchestrateclient.OrchestrateClient

	logger *log.Logger

	ec ethclient.Client

	mux *sync.RWMutex

	TearDownFunc []func()
}

func NewScenarioContext(
	consumerTracker *testutils.ConsumerTracker,
	httpClient *gohttp.Client,
	client orchestrateclient.OrchestrateClient,
	aliasesReg *alias.Registry,
	ec ethclient.Client,
	topics *sarama.TopicConfig,
	waitFor time.Duration,
) *ScenarioContext {
	sc := &ScenarioContext{
		consumerTracker: consumerTracker,
		httpClient:      httpClient,
		aliases:         aliasesReg,
		client:          client,
		logger:          log.NewLogger().SetComponent("e2e.cucumber"),
		ec:              ec,
		topics:          topics,
		waitForEnvelope: waitFor,
		mux:             &sync.RWMutex{},
	}

	return sc
}

// initScenarioContext initialize a scenario context - create a random scenario id - initialize a logger enrich with the scenario name - initialize envelope chan
func (sc *ScenarioContext) init(s *gherkin.Pickle) {
	// Hook the Pickle to the scenario context
	sc.Pickle = s
	sc.aliases.Set(sc.Pickle.Id, sc.Pickle.Id, "scenarioID")

	// Enrich logger
	sc.logger = sc.logger.WithField("scenario.name", sc.Pickle.Name).WithField("scenario.id", sc.Pickle.Id)
}

type stepTable func(*gherkin.PickleStepArgument_PickleTable) error

func (sc *ScenarioContext) preProcessTableStep(tableFunc stepTable) stepTable {
	return func(table *gherkin.PickleStepArgument_PickleTable) error {
		err := sc.replaceAliases(table)
		if err != nil {
			return err
		}

		c, _ := copystructure.Copy(table)
		copyTable := c.(*gherkin.PickleStepArgument_PickleTable)

		return tableFunc(copyTable)
	}
}

func InitializeScenario(s *godog.ScenarioContext, consumerTracker *testutils.ConsumerTracker, topics *sarama.TopicConfig, waitFor time.Duration) {
	sc := NewScenarioContext(
		consumerTracker,
		http.NewClient(http.NewDefaultConfig()),
		orchestrateclient.GlobalClient(),
		alias.GlobalAliasRegistry(),
		rpcClient.GlobalClient(),
		topics,
		waitFor,
	)

	s.BeforeScenario(sc.init)
	s.AfterScenario(sc.tearDown)

	s.BeforeStep(func(s *gherkin.Pickle_PickleStep) {
		sc.logger.WithField("step", s.Text).Debug("step starts")
	})
	s.AfterStep(func(s *gherkin.Pickle_PickleStep, err error) {
		sc.logger.WithField("step", s.Text).Debug("step completed")
	})

	initEnvelopeSteps(s, sc)
	initHTTPSteps(s, sc)
	initContractRegistrySteps(s, sc)
}
