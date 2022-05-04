// +build integration

package integrationtests

import (
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// schedulesTestSuite is a test suite for Schedules controller
type schedulesTestSuite struct {
	suite.Suite
	client sdk.OrchestrateClient
	env    *IntegrationEnvironment
}

func (s *schedulesTestSuite) TestSuccess() {
	ctx := s.env.ctx

	s.T().Run("should create a schedule successfully", func(t *testing.T) {
		schedule, err := s.client.CreateSchedule(ctx, nil)

		require.NoError(t, err)
		assert.NotEmpty(t, schedule.UUID)
		assert.NotEmpty(t, schedule.CreatedAt)
		assert.Equal(t, multitenancy.DefaultTenant, schedule.TenantID)
	})
}

