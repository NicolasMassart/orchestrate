// +build integration

package integrationtests

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/require"

	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// jobsTestSuite is a test suite for Jobs controller
type jobsTestSuite struct {
	suite.Suite
	client    client.OrchestrateClient
	env       *IntegrationEnvironment
	chainUUID string
}

func (s *jobsTestSuite) TestCreate() {
	ctx := s.env.ctx
	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	s.T().Run("should create a new job successfully", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chainUUID
		req.Transaction.From = nil
		req.Annotations = api.Annotations{
			OneTimeKey: true,
		}

		job, err := s.client.CreateJob(ctx, req)
		require.NoError(t, err)

		assert.NotEmpty(t, job.UUID)
		assert.Equal(t, req.Type, job.Type)
		assert.Equal(t, req.ScheduleUUID, job.ScheduleUUID)
		assert.Equal(t, multitenancy.DefaultTenant, job.TenantID)
		assert.Equal(t, s.chainUUID, job.ChainUUID)
		assert.Equal(t, entities.StatusCreated, job.Status)
		assert.Empty(t, job.ParentJobUUID)
		assert.Empty(t, job.NextJobUUID)
		assert.NotEmpty(t, job.CreatedAt)
		assert.NotEmpty(t, job.UpdatedAt)
	})

	s.T().Run("should fail with 400 if type is invalid", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.Type = ""

		_, err := s.client.CreateJob(ctx, req)
		assert.Error(t, err)
		assert.True(t, errors.IsInvalidFormatError(err))
	})

	s.T().Run("should fail with 422 if chainUUID does not exist", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()

		_, err := s.client.CreateJob(ctx, req)
		assert.True(t, errors.IsInvalidParameterError(err))
	})

	s.T().Run("should fail with 422 if schedule does not exit", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ChainUUID = s.chainUUID

		_, err := s.client.CreateJob(ctx, req)
		assert.True(t, errors.IsInvalidParameterError(err))
	})

	s.T().Run("should fail with DependencyFailure if postgres is down", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()

		err := s.env.client.Stop(ctx, postgresContainerID)
		require.NoError(t, err)

		_, err = s.client.CreateJob(ctx, req)
		assert.Error(t, err)

		err = s.env.client.StartServiceAndWait(ctx, postgresContainerID, 10*time.Second)
		assert.NoError(t, err)
	})
}

func (s *jobsTestSuite) TestGet() {
	ctx := s.env.ctx
	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)
	req := testdata.FakeCreateJobRequest()
	req.ScheduleUUID = schedule.UUID
	req.ChainUUID = s.chainUUID
	req.Transaction.From = nil

	origJob, err := s.client.CreateJob(ctx, req)
	require.NoError(s.T(), err)

	s.T().Run("should get a new job successfully", func(t *testing.T) {
		job, err := s.client.GetJob(ctx, origJob.UUID)
		require.NoError(s.T(), err)
		assert.Equal(t, req.Type, job.Type)
		assert.Equal(t, req.ScheduleUUID, job.ScheduleUUID)
	})

	s.T().Run("should fail with DependencyFailure if postgres is down", func(t *testing.T) {
		err := s.env.client.Stop(ctx, postgresContainerID)
		require.NoError(t, err)

		_, err = s.client.GetJob(ctx, origJob.UUID)
		require.Error(t, err)

		err = s.env.client.StartServiceAndWait(ctx, postgresContainerID, 10*time.Second)
		assert.NoError(t, err)
	})
}

func (s *jobsTestSuite) TestStart() {
	ctx := s.env.ctx
	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	s.T().Run("should start a new job successfully", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chainUUID
		req.Transaction.From = nil
		req.Annotations = api.Annotations{
			OneTimeKey: true,
		}

		job, err := s.client.CreateJob(ctx, req)
		require.NoError(t, err)

		err = s.client.StartJob(ctx, job.UUID)
		require.NoError(t, err)
		msgJob, err := s.env.consumer.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Sender, waitForEnvelopeTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(t, msgJob.UUID, job.UUID)

		jobRetrieved, err := s.client.GetJob(ctx, job.UUID)
		require.NoError(t, err)

		assert.Equal(t, entities.StatusStarted, jobRetrieved.Status)
	})

	s.T().Run("should fail with 409 to start if job has already started", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chainUUID
		req.Transaction.From = nil
		req.Annotations = api.Annotations{
			OneTimeKey: true,
		}

		job, err := s.client.CreateJob(ctx, req)
		require.NoError(t, err)

		err = s.client.StartJob(ctx, job.UUID)
		require.NoError(t, err)
		msgJob, err := s.env.consumer.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Sender, waitForEnvelopeTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(t, msgJob.UUID, job.UUID)
	})
}

func (s *jobsTestSuite) TestUpdate() {
	ctx := s.env.ctx
	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	s.T().Run("should update job to MINED and notify", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chainUUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		err = s.client.StartJob(ctx, job.UUID)
		require.NoError(s.T(), err)
		msgJob, err := s.env.consumer.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Sender, waitForEnvelopeTimeOut)
		require.NoError(s.T(), err)
		
		assert.Equal(s.T(), msgJob.UUID, job.UUID)

		receipt := testdata2.FakeReceipt()
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusMined,
			Receipt: receipt,
		})
		require.NoError(s.T(), err)
		txResp, err := s.env.consumer.WaitForTxResponseInTxDecoded(ctx, job.ScheduleUUID, waitForEnvelopeTimeOut)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), txResp.GetJobUUID(), job.UUID)
		assert.Equal(s.T(), txResp.GetReceipt().TxHash, receipt.TxHash)
	})
	
	s.T().Run("should update job to FAILED and notify", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chainUUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		err = s.client.StartJob(ctx, job.UUID)
		require.NoError(s.T(), err)
		msgJob, err := s.env.consumer.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Sender, waitForEnvelopeTimeOut)
		require.NoError(s.T(), err)
		
		assert.Equal(s.T(), msgJob.UUID, job.UUID)

		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusFailed,
		})

		require.NoError(s.T(), err)
		txResp, err := s.env.consumer.WaitForTxResponseInTxRecover(ctx, job.ScheduleUUID, waitForEnvelopeTimeOut)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), txResp.GetJobUUID(), job.UUID)
	})
}
