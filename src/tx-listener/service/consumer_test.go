// +build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	mock3 "github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mocks3 "github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type messageListenerCtrlTestSuite struct {
	suite.Suite
	consumerHandler     *messageConsumerHandler
	pendingJobUC        *mocks.MockPendingJob
	failedJobUC         *mocks.MockFailedJob
	chainSessionMngr    *mocks3.MockChainSessionManager
	retryJobSessionMngr *mocks3.MockRetryJobSessionManager
	apiClient           *mock3.MockOrchestrateClient
	tenantID            string
	allowedTenants      []string
}

func TestMessageListener(t *testing.T) {
	s := new(messageListenerCtrlTestSuite)
	suite.Run(t, s)
}

func (s *messageListenerCtrlTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.tenantID = "tenantID"
	s.allowedTenants = []string{s.tenantID, "_"}
	s.pendingJobUC = mocks.NewMockPendingJob(ctrl)
	s.failedJobUC = mocks.NewMockFailedJob(ctrl)
	s.chainSessionMngr = mocks3.NewMockChainSessionManager(ctrl)
	s.retryJobSessionMngr = mocks3.NewMockRetryJobSessionManager(ctrl)
	s.apiClient = mock3.NewMockOrchestrateClient(ctrl)

	bckoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond*100), 2)
	s.consumerHandler = newMessageConsumerHandler(s.pendingJobUC, s.failedJobUC, s.chainSessionMngr, s.retryJobSessionMngr, bckoff)
}

func (s *messageListenerCtrlTestSuite) TestMessageListener_PublicEthereum() {
	s.T().Run("should handle new pending job successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = json.Marshal(job)

		s.pendingJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), job).Return(nil)
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(nil)

		err := s.consumerHandler.ProcessMsg(context.Background(), msg, job)
		require.NoError(t, err)
	})

	s.T().Run("should handle new pending job ignoring already exiting retry sessions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = json.Marshal(job)

		s.pendingJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), job).Return(errors.AlreadyExistsError(""))
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(nil)

		err := s.consumerHandler.ProcessMsg(context.Background(), msg, job)
		require.NoError(t, err)
	})

	s.T().Run("should handle new pending job ignoring already exiting chain listening sessions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = json.Marshal(job)

		s.pendingJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), job).Return(nil)
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(errors.AlreadyExistsError(""))

		err := s.consumerHandler.ProcessMsg(context.Background(), msg, job)
		require.NoError(t, err)
	})

	s.T().Run("should fail with same error if start chain listening fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = json.Marshal(job)

		expectedErr := fmt.Errorf("failed to start chain session")
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(expectedErr)
		s.failedJobUC.EXPECT().Execute(gomock.Any(), job, expectedErr.Error()).Return(nil)

		err := s.consumerHandler.ProcessMsg(context.Background(), msg, job)
		assert.NoError(t, err)
	})

	s.T().Run("should fail with same error if retry job session fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = json.Marshal(job)

		expectedErr := fmt.Errorf("failed to start chain session")
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), job).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), job).Return(expectedErr)
		s.failedJobUC.EXPECT().Execute(gomock.Any(), job, expectedErr.Error()).Return(nil)

		err := s.consumerHandler.ProcessMsg(context.Background(), msg, job)
		assert.NoError(t, err)
	})

	s.T().Run("should fail with same error if retry job session fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = json.Marshal(job)

		expectedErr := fmt.Errorf("failed to start chain session")
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), job).Return(expectedErr)
		s.failedJobUC.EXPECT().Execute(gomock.Any(), job, expectedErr.Error()).Return(nil)

		err := s.consumerHandler.ProcessMsg(context.Background(), msg, job)
		assert.NoError(t, err)
	})
}