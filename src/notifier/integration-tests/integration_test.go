//go:build integration
// +build integration

package integrationtests

import (
	"context"
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/sdk/messenger"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/consensys/orchestrate/tests/pkg/integration-test"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type apiTestSuite struct {
	suite.Suite
	env       *IntegrationEnvironment
	messenger sdk.MessengerNotifier
}

func (s *apiTestSuite) SetupSuite() {
	err := integrationtest.StartEnvironment(s.env.ctx, s.env)
	require.NoError(s.T(), err)

	kafkaProducer, err := sarama.NewProducer(s.env.cfg.Kafka)
	require.NoError(s.T(), err)

	s.messenger = messenger.NewProducerClient(&messenger.Config{
		TopicTxNotifier: s.env.cfg.ConsumerTopic,
	}, kafkaProducer)

	s.env.logger.Info("setup test suite has completed")
}

func (s *apiTestSuite) TearDownSuite() {
	s.env.Teardown(context.Background())
}

func TestNotifier(t *testing.T) {
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

func (s *apiTestSuite) TestAPI_EventStreams() {
	testSuite := new(notificationsTestSuite)
	testSuite.env = s.env
	testSuite.messenger = s.messenger
	suite.Run(s.T(), testSuite)
}
