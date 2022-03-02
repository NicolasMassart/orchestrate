// +build integration

package integrationtests

import (
	"context"
	"os"
	"testing"
	"time"

	integrationtest "github.com/consensys/orchestrate/pkg/integration-test"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type chainListenerTestSuite struct {
	suite.Suite
	env *IntegrationEnvironment
	err error
}

func (s *chainListenerTestSuite) SetupSuite() {
	err := integrationtest.StartEnvironment(s.env.ctx, s.env)
	if err != nil {
		s.env.logger.WithError(err).Error()
		if s.err == nil {
			s.err = err
		}
		return
	}

	s.env.logger.Infof("setup test suite has completed")
}

func (s *chainListenerTestSuite) TearDownSuite() {
	s.env.Teardown(context.Background())

	if s.err != nil {
		s.Fail(s.err.Error())
	}
}

func TestChainListener(t *testing.T) {
	s := new(chainListenerTestSuite)
	ctx, cancel := context.WithCancel(context.Background())

	s.env, s.err = NewIntegrationEnvironment(ctx, cancel)
	if s.err != nil {
		t.Errorf(s.err.Error())
		return
	}

	sig := utils.NewSignalListener(func(signal os.Signal) {
		cancel()
	})
	defer sig.Close()

	suite.Run(t, s)
}

func (s *chainListenerTestSuite) TestTxListener() {
	if s.err != nil {
		s.env.logger.Warn("skipping test...")
		return
	}

	testSuite := new(txListenerTestSuite)
	testSuite.env = s.env

	time.Sleep(3 * time.Second)
	suite.Run(s.T(), testSuite)
}
