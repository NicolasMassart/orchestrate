// +build integration

package integrationtests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type txListenerTestSuite struct {
	suite.Suite
	env *IntegrationEnvironment
}

func (s *txListenerTestSuite) TestEmpty() {
	s.T().Run("should hello world example", func(t *testing.T) {
		
	})
}
