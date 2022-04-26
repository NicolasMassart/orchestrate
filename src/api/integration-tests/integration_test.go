//go:build integration
// +build integration

package integrationtests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	integrationtest "github.com/consensys/orchestrate/tests/pkg/integration-test"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type apiTestSuite struct {
	suite.Suite
	env    *IntegrationEnvironment
	client client.OrchestrateClient
}

func (s *apiTestSuite) SetupSuite() {
	err := integrationtest.StartEnvironment(s.env.ctx, s.env)
	require.NoError(s.T(), err)
	time.Sleep(1 * time.Second)

	s.env.logger.Debug("setting up test accounts and chains")

	conf := client.NewConfig(s.env.baseURL, "", nil)
	conf.MetricsURL = s.env.metricsURL
	s.client = client.NewHTTPClient(http.NewClient(http.NewDefaultConfig()), conf)

	// @TODO Remove this hidden dependency
	// We use this chain in the API_Transactions tests
	_, err = s.client.RegisterChain(s.env.ctx, &api.RegisterChainRequest{
		Name: "ganache",
		URLs: []string{s.env.blockchainNodeURL},
	})
	require.NoError(s.T(), err)

	// We use this account in the API_Transactions and Faucet tests
	_, err = s.client.ImportAccount(s.env.ctx, testdata.FakeImportAccountRequest())
	require.NoError(s.T(), err)

	s.env.logger.Info("setup test suite has completed")
}

func (s *apiTestSuite) TearDownSuite() {
	s.env.Teardown(context.Background())
}

func TestAPI(t *testing.T) {
	s := new(apiTestSuite)
	ctx, cancel := context.WithCancel(context.Background())

	env, err := NewIntegrationEnvironment(ctx)
	require.NoError(t, err)
	s.env = env

	sig := utils.NewSignalListener(func(signal os.Signal) {
		cancel()
	})
	defer sig.Close()

	suite.Run(t, s)
}

func (s *apiTestSuite) TestAPI_Transactions() {
	testSuite := new(transactionsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Accounts() {
	testSuite := new(accountsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	testSuite.defaultQKMStoreID = qkmDefaultStoreID
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Jobs() {
	testSuite := new(jobsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Schedules() {
	testSuite := new(schedulesTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Contracts() {
	testSuite := new(contractsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Chains() {
	testSuite := new(chainsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Faucets() {
	testSuite := new(faucetsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_EventStreams() {
	testSuite := new(eventStreamsTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

func (s *apiTestSuite) TestAPI_Proxy() {
	testSuite := new(proxyTestSuite)
	testSuite.env = s.env
	testSuite.client = s.client
	suite.Run(s.T(), testSuite)
}

// func (s *apiTestSuite) TestAPI_Metrics() {
// 	if s.err != nil {
// 		s.env.logger.Warn("skipping test...")
// 		return
// 	}
//
// 	testSuite := new(metricsTestSuite)
// 	testSuite.env = s.env
// 	testSuite.client = s.client
// 	suite.Run(s.T(), testSuite)
// }
