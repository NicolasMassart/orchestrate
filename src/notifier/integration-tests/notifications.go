//go:build integration
// +build integration

package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities/testdata"
	notifierTypes "github.com/consensys/orchestrate/src/notifier/service/types"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const waitForNotificationTimeOut = 5 * time.Second

type notificationsTestSuite struct {
	suite.Suite
	messenger sdk.MessengerNotifier
	env       *IntegrationEnvironment
}

func (s *notificationsTestSuite) TestSendKafka() {
	ctx := s.env.ctx

	es := testdata.FakeKafkaEventStream()
	es.Kafka.Topic = s.env.notificationTopic
	userInfo := multitenancy.NewInternalAdminUser()

	s.T().Run("should send notification successfully: MINED", func(t *testing.T) {
		notif := testdata.FakeNotification()
		notif.Type = entities.NotificationTypeTxMined
		notif.Error = ""

		err := s.messenger.TransactionNotificationMessage(ctx, es, notif, userInfo)
		require.NoError(t, err)

		notificationRes, err := s.env.notifierConsumerTracker.WaitForMinedTransaction(ctx, notif.SourceUUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationRes.SourceUUID, notif.SourceUUID)
		assert.Equal(s.T(), notificationRes.Status, notif.Status.String())
		assert.Equal(s.T(), notificationRes.Type, notif.Type.String())
		assert.Equal(s.T(), notificationRes.UpdatedAt, notif.UpdatedAt)
		assert.Equal(s.T(), notificationRes.CreatedAt, notif.CreatedAt)
		assert.Equal(s.T(), notificationRes.APIVersion, notif.APIVersion)
		assert.Equal(s.T(), notificationRes.UUID, notif.UUID)

		// TODO: Improve tests when txResponse formatting is done
		assert.Equal(s.T(), notificationRes.Data.(*entities.Job).UUID, notif.Job.UUID)

		notificationReq, err := s.env.messengerConsumerTracker.WaitForAckNotif(ctx, notif.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationReq.UUID, notif.UUID)
	})

	s.T().Run("should send notification successfully: FAILED", func(t *testing.T) {
		notif := testdata.FakeNotification()

		err := s.messenger.TransactionNotificationMessage(ctx, es, notif, userInfo)
		require.NoError(t, err)

		notificationRes, err := s.env.notifierConsumerTracker.WaitForFailedTransaction(ctx, notif.SourceUUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationRes.SourceUUID, notif.SourceUUID)
		assert.Equal(s.T(), notificationRes.Status, notif.Status.String())
		assert.Equal(s.T(), notificationRes.Type, notif.Type.String())
		assert.Equal(s.T(), notificationRes.UpdatedAt, notif.UpdatedAt)
		assert.Equal(s.T(), notificationRes.CreatedAt, notif.CreatedAt)
		assert.Equal(s.T(), notificationRes.APIVersion, notif.APIVersion)
		assert.Equal(s.T(), notificationRes.UUID, notif.UUID)
		assert.Equal(s.T(), notificationRes.Error, notif.Error)
		assert.Nil(s.T(), notificationRes.Data)

		notificationReq, err := s.env.messengerConsumerTracker.WaitForAckNotif(ctx, notif.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationReq.UUID, notif.UUID)
	})
}

func (s *notificationsTestSuite) TestSendWebhook() {
	ctx := s.env.ctx

	es := testdata.FakeWebhookEventStream()
	webhookDomainURL := "http://webhook.com"
	webhookURLPath := "/wait-for-notification"
	es.Webhook.URL = webhookDomainURL + webhookURLPath
	userInfo := multitenancy.NewInternalAdminUser()

	s.T().Run("should send notification successfully: MINED", func(t *testing.T) {
		notif := testdata.FakeNotification()
		notif.Type = entities.NotificationTypeTxMined
		notif.Error = ""

		waitNotification := make(chan *notifierTypes.NotificationResponse, 1)
		waitNotificationErr := make(chan error, 1)
		gock.New(webhookDomainURL).Post(webhookURLPath).
			AddMatcher(webhookCallMatcher(waitNotification, waitNotificationErr, waitForNotificationTimeOut)).
			Reply(http.StatusOK)

		err := s.messenger.TransactionNotificationMessage(ctx, es, notif, userInfo)
		require.NoError(t, err)

		select {
		case notificationRes := <-waitNotification:
			assert.Equal(s.T(), notificationRes.SourceUUID, notif.SourceUUID)
			assert.Equal(s.T(), notificationRes.Type, string(entities.NotificationTypeTxMined))
			assert.Equal(s.T(), notificationRes.Data.(map[string]interface{})["UUID"].(string), notif.Job.UUID)
		case err := <-waitNotificationErr:
			assert.Error(t, err)
		}

		notificationReq, err := s.env.messengerConsumerTracker.WaitForAckNotif(ctx, notif.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationReq.UUID, notif.UUID)

	})

	s.T().Run("should update job to FAILED and notify", func(t *testing.T) {
		notif := testdata.FakeNotification()

		waitNotification := make(chan *notifierTypes.NotificationResponse, 1)
		waitNotificationErr := make(chan error, 1)
		gock.New(webhookDomainURL).Post(webhookURLPath).
			AddMatcher(webhookCallMatcher(waitNotification, waitNotificationErr, waitForNotificationTimeOut)).
			Reply(http.StatusOK)

		err := s.messenger.TransactionNotificationMessage(ctx, es, notif, userInfo)
		require.NoError(t, err)

		select {
		case notification := <-waitNotification:
			assert.Equal(s.T(), notification.SourceUUID, notif.SourceUUID)
			assert.Equal(s.T(), notification.Error, notif.Error)
			assert.Equal(s.T(), notification.Type, string(entities.NotificationTypeTxFailed))
			assert.Nil(s.T(), notification.Data)
		case err := <-waitNotificationErr:
			assert.Error(t, err)
		}

		notificationReq, err := s.env.messengerConsumerTracker.WaitForAckNotif(ctx, notif.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationReq.UUID, notif.UUID)
	})
}

func (s *notificationsTestSuite) TestSuspend() {
	ctx := s.env.ctx

	webhookDomainURL := "http://webhook.com"
	userInfo := multitenancy.NewInternalAdminUser()

	s.T().Run("should suspend event stream if it fails to send notification and process next one", func(t *testing.T) {
		es0 := testdata.FakeWebhookEventStream()
		webhookURLPath := "/inexistent-path"
		es0.Webhook.URL = webhookDomainURL + webhookURLPath

		err := s.messenger.TransactionNotificationMessage(ctx, es0, testdata.FakeNotification(), userInfo)
		require.NoError(t, err)

		es1 := testdata.FakeWebhookEventStream()
		webhookURLPath = "/existent-path"
		es1.Webhook.URL = webhookDomainURL + webhookURLPath
		expectedNotif := testdata.FakeNotification()

		err = s.messenger.TransactionNotificationMessage(ctx, es1, expectedNotif, userInfo)
		require.NoError(t, err)

		waitNotification := make(chan *notifierTypes.NotificationResponse, 1)
		waitNotificationErr := make(chan error, 1)
		gock.New(webhookDomainURL).Post(webhookURLPath).
			AddMatcher(webhookCallMatcher(waitNotification, waitNotificationErr, waitForNotificationTimeOut)).
			Reply(http.StatusOK)

		notificationRes, err := s.env.messengerConsumerTracker.WaitForSuspendEventStream(ctx, es0.UUID, waitForNotificationTimeOut)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), notificationRes.UUID, es0.UUID)

		select {
		case notification := <-waitNotification:
			assert.Equal(s.T(), notification.SourceUUID, expectedNotif.SourceUUID)
			assert.Equal(s.T(), notification.Error, expectedNotif.Error)
			assert.Equal(s.T(), notification.Type, string(entities.NotificationTypeTxFailed))
			assert.Nil(s.T(), notification.Data)
		case err := <-waitNotificationErr:
			assert.Error(t, err)
		}
	})
}

func webhookCallMatcher(cNotification chan *notifierTypes.NotificationResponse, cErr chan error, duration time.Duration) gock.MatchFunc {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	go func() {
		<-ticker.C
		cErr <- fmt.Errorf("timeout after %s", duration.String())
	}()

	return func(rw *http.Request, grw *gock.Request) (bool, error) {
		body, _ := ioutil.ReadAll(rw.Body)
		rw.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		req := &notifierTypes.NotificationResponse{}
		if err := json.Unmarshal(body, &req); err != nil {
			cErr <- err
			return false, err
		}

		cNotification <- req
		return true, nil
	}
}
