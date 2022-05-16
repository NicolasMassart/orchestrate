//go:build integration
// +build integration

package integrationtests

import (
	"net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type jobsTestSuite struct {
	suite.Suite
	client sdk.OrchestrateClient
	env    *IntegrationEnvironment
	chain  *api.ChainResponse
}

func (s *jobsTestSuite) SetupSuite() {
	ctx := s.env.ctx

	chainReq := testdata.FakeRegisterChainRequest()
	chainReq.URLs = []string{s.env.blockchainNodeURL}
	chainReq.PrivateTxManagerURL = ""

	var err error
	s.chain, err = s.client.RegisterChain(ctx, chainReq)
	require.NoError(s.T(), err)
}

func (s *jobsTestSuite) TestCreate() {
	ctx := s.env.ctx
	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	s.T().Run("should create a new job successfully", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
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
		assert.Equal(t, s.chain.UUID, job.ChainUUID)
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
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 422 if chain does not exist", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()

		_, err := s.client.CreateJob(ctx, req)
		assert.Equal(t, http.StatusUnprocessableEntity, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail with 422 if schedule does not exit", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ChainUUID = s.chain.UUID

		_, err := s.client.CreateJob(ctx, req)
		assert.Equal(t, http.StatusUnprocessableEntity, err.(*client.HTTPErr).Code())
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
	req.ChainUUID = s.chain.UUID
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
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		req.Annotations = api.Annotations{
			OneTimeKey: true,
		}

		job, err := s.client.CreateJob(ctx, req)
		require.NoError(t, err)

		err = s.client.StartJob(ctx, job.UUID)
		require.NoError(t, err)
		msg, err := s.env.messengerConsumerTracker.WaitForStartedJobMessage(ctx, job.UUID, waitForNotificationTimeOut)
		require.NoError(t, err)
		assert.Equal(t, msg.Job.UUID, job.UUID)

		jobRetrieved, err := s.client.GetJob(ctx, job.UUID)
		require.NoError(t, err)

		assert.Equal(t, entities.StatusStarted, jobRetrieved.Status)
	})

	s.T().Run("should fail with 409 to start if job has already started", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		req.Annotations = api.Annotations{
			OneTimeKey: true,
		}

		job, err := s.client.CreateJob(ctx, req)
		require.NoError(t, err)

		err = s.client.StartJob(ctx, job.UUID)
		require.NoError(t, err)

		msgJob, err := s.env.messengerConsumerTracker.WaitForStartedJobMessage(ctx, job.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)
		assert.Equal(t, msgJob.Job.UUID, job.UUID)
	})
}

func (s *jobsTestSuite) TestUpdatePending() {
	ctx := s.env.ctx

	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	s.T().Run("should update job to PENDING and send message", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status: entities.StatusStarted,
		})

		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status: entities.StatusPending,
		})
		require.NoError(s.T(), err)

		_, err = s.env.messengerConsumerTracker.WaitForPendingJobMessage(ctx, job.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)
	})

}

func (s *jobsTestSuite) TestUpdateNotifyWithKafka() {
	ctx := s.env.ctx

	eventStream, err := s.client.CreateEventStream(ctx, &api.CreateEventStreamRequest{
		Channel: "kafka",
		Name:    "integration-test-event-stream-kafka",
		Kafka: &api.KafkaRequest{
			Topic: s.env.notificationTopic,
		},
		Chain: "*",
	})
	require.NoError(s.T(), err)

	defer func() {
		err := s.client.DeleteEventStream(ctx, eventStream.UUID)
		assert.NoError(s.T(), err)
	}()

	s.T().Run("should update job to MINED and notify", func(t *testing.T) {
		schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
		require.NoError(s.T(), err)

		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status: entities.StatusStarted,
		})

		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status: entities.StatusPending,
		})

		receipt := testdata2.FakeReceipt()
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusMined,
			Receipt: receipt,
		})
		require.NoError(s.T(), err)

		notificationReq, err := s.env.messengerConsumerTracker.WaitForTransactionNotificationMessage(ctx, job.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationReq.Notification.Job.UUID, job.UUID)
		assert.Equal(s.T(), notificationReq.EventStream.UUID, eventStream.UUID)
	})

	s.T().Run("should update job to FAILED and notify", func(t *testing.T) {
		schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
		require.NoError(s.T(), err)

		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status: entities.StatusStarted,
		})

		failedErrMsg := "ErrorMsg"
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusFailed,
			Message: failedErrMsg,
		})
		require.NoError(s.T(), err)

		notificationReq, err := s.env.messengerConsumerTracker.WaitForTransactionNotificationMessage(ctx, job.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationReq.Notification.SourceUUID, job.ScheduleUUID)
		assert.Equal(s.T(), notificationReq.EventStream.UUID, eventStream.UUID)
		assert.Equal(s.T(), notificationReq.Notification.Error, failedErrMsg)
	})
}
