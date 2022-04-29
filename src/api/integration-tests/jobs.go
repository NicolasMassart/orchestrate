//go:build integration
// +build integration

package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/consensys/orchestrate/src/infra/messenger/types"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type jobsTestSuite struct {
	suite.Suite
	client client.OrchestrateClient
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
		msgJob, err := s.env.messengerConsumerTracker.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Sender, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(t, msgJob.UUID, job.UUID)

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
		msgJob, err := s.env.messengerConsumerTracker.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Sender, waitForNotificationTimeOut)
		require.NoError(s.T(), err)
		assert.Equal(t, msgJob.UUID, job.UUID)
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
			Status: entities.StatusPending,
		})
		require.NoError(s.T(), err)

		_, err = s.env.messengerConsumerTracker.WaitForJob(ctx, job.UUID, s.env.apiCfg.KafkaTopics.Listener, waitForNotificationTimeOut)
		require.NoError(s.T(), err)
	})

}

func (s *jobsTestSuite) TestUpdateNotifyWithKafka() {
	ctx := s.env.ctx

	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	eventStream, err := s.client.CreateKafkaEventStream(ctx, &api.CreateKafkaEventStreamRequest{
		Name:  "integration-test-event-stream-kafka",
		Topic: s.env.notificationTopic,
		Chain: s.chain.Name,
	})
	require.NoError(s.T(), err)

	defer func() {
		err := s.client.DeleteEventStream(ctx, eventStream.UUID)
		require.NoError(s.T(), err)
	}()

	s.T().Run("should update job to MINED and notify", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		receipt := testdata2.FakeReceipt()
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusMined,
			Receipt: receipt,
		})
		require.NoError(s.T(), err)

		notification, err := s.env.notifierConsumerTracker.WaitForTxMinedNotification(ctx, job.ScheduleUUID, s.env.notificationTopic, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notification.SourceUUID, job.ScheduleUUID)
		assert.Equal(s.T(), notification.Data.(map[string]interface{})["UUID"].(string), job.UUID)
	})

	s.T().Run("should update job to FAILED and notify", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		failedErrMsg := "ErrorMsg"
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusFailed,
			Message: failedErrMsg,
		})

		notification, err := s.env.notifierConsumerTracker.WaitForTxFailedNotification(ctx, job.ScheduleUUID, s.env.notificationTopic, waitForNotificationTimeOut)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), notification.Error, failedErrMsg)
		assert.Equal(s.T(), notification.SourceUUID, job.ScheduleUUID)
	})
}

func (s *jobsTestSuite) TestUpdateNotifyWithWebhook() {
	ctx := s.env.ctx

	webhookDomainURL := "http://webhook.com"
	webhookURLPath := "/wait-for-notification"

	schedule, err := s.client.CreateSchedule(ctx, &api.CreateScheduleRequest{})
	require.NoError(s.T(), err)

	eventStream, err := s.client.CreateWebhookEventStream(ctx, &api.CreateWebhookEventStreamRequest{
		Name:  "integration-test-event-stream-webhook",
		URL:   webhookDomainURL + webhookURLPath,
		Chain: s.chain.Name,
	})
	require.NoError(s.T(), err)

	defer func() {
		err := s.client.DeleteEventStream(ctx, eventStream.UUID)
		require.NoError(s.T(), err)
	}()

	s.T().Run("should update job to MINED and notify", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		waitNotification := make(chan *types.NotificationResponse, 1)
		waitNotificationErr := make(chan error, 1)
		gock.New(webhookDomainURL).Post(webhookURLPath).
			AddMatcher(webhookCallMatcher(waitNotification, waitNotificationErr, waitForNotificationTimeOut)).
			Reply(http.StatusOK)

		receipt := testdata2.FakeReceipt()
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusMined,
			Receipt: receipt,
		})
		require.NoError(s.T(), err)

		select {
		case notification := <-waitNotification:
			assert.Equal(s.T(), notification.SourceUUID, job.ScheduleUUID)
			assert.Equal(s.T(), notification.Type, string(entities.NotificationTypeTxMined))
			assert.Equal(s.T(), notification.Data.(map[string]interface{})["UUID"].(string), job.UUID)
		case err := <-waitNotificationErr:
			assert.Error(t, err)
		}
	})

	s.T().Run("should update job to FAILED and notify", func(t *testing.T) {
		req := testdata.FakeCreateJobRequest()
		req.ScheduleUUID = schedule.UUID
		req.ChainUUID = s.chain.UUID
		req.Transaction.From = nil
		job, err := s.client.CreateJob(ctx, req)
		require.NoError(s.T(), err)

		waitNotification := make(chan *types.NotificationResponse, 1)
		waitNotificationErr := make(chan error, 1)
		gock.New(webhookDomainURL).Post(webhookURLPath).
			AddMatcher(webhookCallMatcher(waitNotification, waitNotificationErr, waitForNotificationTimeOut)).
			Reply(http.StatusOK)

		failedErrMsg := "errMsg"
		_, err = s.client.UpdateJob(ctx, job.UUID, &api.UpdateJobRequest{
			Status:  entities.StatusFailed,
			Message: failedErrMsg,
		})
		require.NoError(s.T(), err)

		select {
		case notification := <-waitNotification:
			assert.Equal(s.T(), notification.SourceUUID, job.ScheduleUUID)
			assert.Equal(s.T(), notification.Error, failedErrMsg)
			assert.Equal(s.T(), notification.Type, string(entities.NotificationTypeTxFailed))
		case err := <-waitNotificationErr:
			assert.Error(t, err)
		}
	})
}

func webhookCallMatcher(cNotification chan *types.NotificationResponse, cErr chan error, duration time.Duration) gock.MatchFunc {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	go func() {
		<-ticker.C
		cErr <- fmt.Errorf("timeout after %s", duration.String())
	}()

	return func(rw *http.Request, grw *gock.Request) (bool, error) {
		body, _ := ioutil.ReadAll(rw.Body)
		rw.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		req := &types.NotificationResponse{}
		if err := json.Unmarshal(body, &req); err != nil {
			cErr <- err
			return false, err
		}

		cNotification <- req
		return true, nil
	}
}
