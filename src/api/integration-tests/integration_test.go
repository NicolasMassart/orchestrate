// +build integration

package integrationtests

import (
	"context"
	integrationtest "github.com/consensys/orchestrate/pkg/integration-test"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	entitiestestdata "github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
	"os"
	"testing"
	"time"
)

type apiTestSuite struct {
	suite.Suite
	env       *IntegrationEnvironment
	client    client.OrchestrateClient
	chainUUID string
}

func (s *apiTestSuite) SetupSuite() {
	defer gock.Off()

	err := integrationtest.StartEnvironment(s.env.ctx, s.env)
	require.NoError(s.T(), err)
	time.Sleep(1 * time.Second)

	s.env.logger.Debug("setting up test accounts and chains")

	conf := client.NewConfig(s.env.baseURL, "", nil)
	conf.MetricsURL = s.env.metricsURL
	s.client = client.NewHTTPClient(http.NewClient(http.NewDefaultConfig()), conf)

	// We use this chain in the tests
	chain, err := s.client.RegisterChain(s.env.ctx, &api.RegisterChainRequest{
		Name: "ganache",
		URLs: []string{s.env.blockchainNodeURL},
		Listener: api.RegisterListenerRequest{
			FromBlock:         "latest",
			ExternalTxEnabled: false,
		},
	})
	require.NoError(s.T(), err)
	s.chainUUID = chain.UUID

	// We use this account in the tests
	account := entitiestestdata.FakeAccount()
	account.Address = testdata.FromAddress
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
	testSuite.chainUUID = s.chainUUID
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