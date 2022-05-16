package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/mock"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/tx-listener/service/types"
	mocks3 "github.com/consensys/orchestrate/src/tx-listener/tx-listener/sessions/mocks"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type jobHandlerTestSuite struct {
	suite.Suite
	router              *JobHandler
	pendingJobUC        *mocks.MockPendingJob
	failedJobUC         *mocks.MockFailedJob
	chainSessionMngr    *mocks3.MockChainSessionManager
	retryJobSessionMngr *mocks3.MockRetryJobSessionManager
	apiClient           *mock.MockOrchestrateClient
	tenantID            string
	allowedTenants      []string
}

func TestJobHandler(t *testing.T) {
	s := new(jobHandlerTestSuite)
	suite.Run(t, s)
}

func (s *jobHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.tenantID = "tenantID"
	s.allowedTenants = []string{s.tenantID, "_"}
	s.pendingJobUC = mocks.NewMockPendingJob(ctrl)
	s.failedJobUC = mocks.NewMockFailedJob(ctrl)
	s.chainSessionMngr = mocks3.NewMockChainSessionManager(ctrl)
	s.retryJobSessionMngr = mocks3.NewMockRetryJobSessionManager(ctrl)
	s.apiClient = mock.NewMockOrchestrateClient(ctrl)

	bckoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond*100), 2)
	s.router = NewJobHandler(s.pendingJobUC, s.failedJobUC, s.chainSessionMngr, s.retryJobSessionMngr, bckoff)
}

func (s *jobHandlerTestSuite) TestMessageListener_PublicEthereum() {
	s.T().Run("should handle new pending job successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second

		msg := newPendingJobMsg(job, 1)
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), msg).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), testdata.NewJobMatcher(job)).Return(nil)
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(nil)

		err := s.router.HandlePendingJobMessage(context.Background(), msg)
		require.NoError(t, err)
	})

	s.T().Run("should handle new pending job ignoring already exiting retry sessions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := newPendingJobMsg(job, 1)

		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(nil)
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), msg).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), testdata.NewJobMatcher(job)).Return(errors.AlreadyExistsError(""))

		err := s.router.HandlePendingJobMessage(context.Background(), msg)
		require.NoError(t, err)
	})

	s.T().Run("should handle new pending job ignoring already exiting chain listening sessions", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := newPendingJobMsg(job, 1)

		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(errors.AlreadyExistsError(""))
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), msg).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), testdata.NewJobMatcher(job)).Return(nil)

		err := s.router.HandlePendingJobMessage(context.Background(), msg)
		require.NoError(t, err)
	})

	s.T().Run("should fail with same error if start chain listening fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		msg := newPendingJobMsg(job, 1)

		expectedErr := fmt.Errorf("failed to start chain session")
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(expectedErr)
		s.failedJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), expectedErr.Error()).Return(nil)

		err := s.router.HandlePendingJobMessage(context.Background(), msg)
		assert.NoError(t, err)
	})

	s.T().Run("should fail with same error if retry job session fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		job.InternalData.RetryInterval = time.Second
		msg := newPendingJobMsg(job, 1)

		expectedErr := fmt.Errorf("failed to start chain session")
		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(nil)
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), msg).Return(nil)
		s.retryJobSessionMngr.EXPECT().StartSession(gomock.Any(), testdata.NewJobMatcher(job)).Return(expectedErr)
		s.failedJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), expectedErr.Error()).Return(nil)

		err := s.router.HandlePendingJobMessage(context.Background(), msg)
		assert.NoError(t, err)
	})

	s.T().Run("should fail with same error if retry job session fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.TenantID = s.tenantID
		msg := newPendingJobMsg(job, 2)

		s.chainSessionMngr.EXPECT().StartSession(gomock.Any(), job.ChainUUID).Return(nil)
		expectedErr := fmt.Errorf("failed to start retry session")
		s.pendingJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), msg).Return(expectedErr)
		s.failedJobUC.EXPECT().Execute(gomock.Any(), testdata.NewJobMatcher(job), expectedErr.Error()).Return(nil)

		err := s.router.HandlePendingJobMessage(context.Background(), msg)
		assert.NoError(t, err)
	})
}

func newPendingJobMsg(job *entities.Job, offset int64) *entities.Message {
	bMsgBody, _ := json.Marshal(&types.PendingJobMessageRequest{
		Job: job,
	})
	return &entities.Message{
		Type:   PendingJobMessageType,
		Body:   bMsgBody,
		Offset: offset,
	}
}
