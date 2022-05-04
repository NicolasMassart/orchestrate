// +build unit

package service
// 
// import (
// 	"context"
// 	"encoding/json"
// 	"testing"
// 	"time"
// 
// 	"github.com/Shopify/sarama"
// 	"github.com/cenkalti/backoff/v4"
// 	"github.com/consensys/orchestrate/pkg/errors"
// 	mock3 "github.com/consensys/orchestrate/pkg/sdk/client/mock"
// 	"github.com/consensys/orchestrate/pkg/sdk/mock"
// 	api "github.com/consensys/orchestrate/src/api/service/types"
// 	"github.com/consensys/orchestrate/src/entities"
// 	"github.com/consensys/orchestrate/src/entities/testdata"
// 	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
// 	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/mocks"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// )
// 
// const errMsgExceedTime = "exceeded waiting time"
// 
// type messageListenerCtrlTestSuite struct {
// 	suite.Suite
// 	consumerHandler     *Router
// 	sendETHUC           *mocks.MockSendETHTxUseCase
// 	sendETHRawUC        *mocks.MockSendETHRawTxUseCase
// 	sendEEAPrivateUC    *mocks.MockSendEEAPrivateTxUseCase
// 	sendGoQuorumMarking *mocks.MockSendGoQuorumMarkingTxUseCase
// 	sendGoQuorumPrivate *mocks.MockSendGoQuorumPrivateTxUseCase
// 	apiClient           *mock.MockOrchestrateClient
// 	tenantID            string
// 	allowedTenants      []string
// }
// 
// var _ usecases.UseCases = &messageListenerCtrlTestSuite{}
// 
// func (s *messageListenerCtrlTestSuite) SendETHRawTx() usecases.SendETHRawTxUseCase {
// 	return s.sendETHRawUC
// }
// 
// func (s *messageListenerCtrlTestSuite) SendETHTx() usecases.SendETHTxUseCase {
// 	return s.sendETHUC
// }
// 
// func (s *messageListenerCtrlTestSuite) SendEEAPrivateTx() usecases.SendEEAPrivateTxUseCase {
// 	return s.sendEEAPrivateUC
// }
// 
// func (s *messageListenerCtrlTestSuite) SendGoQuorumPrivateTx() usecases.SendGoQuorumPrivateTxUseCase {
// 	return s.sendGoQuorumPrivate
// }
// 
// func (s *messageListenerCtrlTestSuite) SendGoQuorumMarkingTx() usecases.SendGoQuorumMarkingTxUseCase {
// 	return s.sendGoQuorumMarking
// }
// 
// func TestMessageListener(t *testing.T) {
// 	s := new(messageListenerCtrlTestSuite)
// 	suite.Run(t, s)
// }
// 
// func (s *messageListenerCtrlTestSuite) SetupTest() {
// 	ctrl := gomock.NewController(s.T())
// 	defer ctrl.Finish()
// 
// 	s.tenantID = "tenantID"
// 	s.allowedTenants = []string{s.tenantID, "_"}
// 	s.sendETHRawUC = mocks.NewMockSendETHRawTxUseCase(ctrl)
// 	s.sendETHUC = mocks.NewMockSendETHTxUseCase(ctrl)
// 	s.sendEEAPrivateUC = mocks.NewMockSendEEAPrivateTxUseCase(ctrl)
// 	s.sendGoQuorumPrivate = mocks.NewMockSendGoQuorumPrivateTxUseCase(ctrl)
// 	s.sendGoQuorumMarking = mocks.NewMockSendGoQuorumMarkingTxUseCase(ctrl)
// 	s.apiClient = mock.NewMockOrchestrateClient(ctrl)
// 
// 	bckoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond*100), 2)
// 	s.consumerHandler = newRouter(s, s.apiClient, bckoff)
// }
// 
// func (s *messageListenerCtrlTestSuite) TestMessageListener_PublicEthereum() {
// 	s.T().Run("should execute use case for multiple public ethereum transactions", func(t *testing.T) {
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 
// 		cjob := make(chan *entities.Job, 1)
// 		s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 			cjob <- job
// 			return nil
// 		})
// 
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			assert.Equal(t, rjob.UUID, jobMsg.UUID)
// 		}
// 	})
// 	
// 	s.T().Run("should execute use case for public raw ethereum transactions", func(t *testing.T) {
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		jobMsg.Type = entities.EthereumRawTransaction
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 	
// 		cjob := make(chan *entities.Job, 1)
// 		s.sendETHRawUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 			cjob <- job
// 			return nil
// 		})
// 	
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 	
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			assert.Equal(t, rjob.UUID, jobMsg.UUID)
// 		}
// 	})
// 	
// 	s.T().Run("should execute use case for eea transactions", func(t *testing.T) {
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		jobMsg.Type = entities.EEAPrivateTransaction
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 	
// 		cjob := make(chan *entities.Job, 1)
// 		s.sendEEAPrivateUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 			cjob <- job
// 			return nil
// 		})
// 	
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 	
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			assert.Equal(t, rjob.UUID, jobMsg.UUID)
// 		}
// 	})
// 	
// 	s.T().Run("should execute use case for tessera marking transactions", func(t *testing.T) {
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		jobMsg.Type = entities.GoQuorumMarkingTransaction
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 	
// 		cjob := make(chan *entities.Job, 1)
// 		s.sendGoQuorumMarking.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 			cjob <- job
// 			return nil
// 		})
// 	
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 	
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			assert.Equal(t, rjob.UUID, jobMsg.UUID)
// 		}
// 	})
// 	
// 	s.T().Run("should execute use case for tessera private transactions", func(t *testing.T) {
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		jobMsg.Type = entities.GoQuorumPrivateTransaction
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 	
// 		cjob := make(chan *entities.Job, 1)
// 		s.sendGoQuorumPrivate.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 			cjob <- job
// 			return nil
// 		})
// 	
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 	
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			assert.Equal(t, rjob.UUID, jobMsg.UUID)
// 		}
// 	})
// }
// 
// func (s *messageListenerCtrlTestSuite) TestMessageListener_PublicEthereum_Errors() {
// 	s.T().Run("should update transaction and send message to tx-recover if sending fails", func(t *testing.T) {
// 		expectedErr := errors.InternalError("error")
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 	
// 		cjob := make(chan *entities.Job, 1)
// 		s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 			cjob <- job
// 			return expectedErr
// 		})
// 		s.apiClient.EXPECT().UpdateJob(gomock.Any(), jobMsg.UUID, &api.UpdateJobRequest{
// 			Status:      entities.StatusFailed,
// 			Message:     expectedErr.Error(),
// 			Transaction: nil,
// 		}).Return(&api.JobResponse{}, nil)
// 	
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 	
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			time.Sleep(time.Millisecond * 500) // Wait for receipt to be sent
// 			assert.Equal(t, jobMsg.UUID, rjob.UUID)
// 		}
// 	})
// 	
// 	s.T().Run("should update transaction and retry job if sending fails by nonce error", func(t *testing.T) {
// 		invalidNonceErr := errors.InvalidNonceWarning("nonce too low")
// 		jobMsg := testdata.FakeJob()
// 		jobMsg.TenantID = s.tenantID
// 		msg := &sarama.ConsumerMessage{}
// 		msg.Value, _ = json.Marshal(jobMsg)
// 	
// 		cjob := make(chan *entities.Job, 1)
// 		gomock.InOrder(
// 			s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(invalidNonceErr),
// 			s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
// 				cjob <- job
// 				return nil
// 			}),
// 		)
// 	
// 		s.apiClient.EXPECT().UpdateJob(gomock.Any(), jobMsg.UUID, &api.UpdateJobRequest{
// 			Status:      entities.StatusRecovering,
// 			Message:     invalidNonceErr.Error(),
// 			Transaction: nil,
// 		}).Return(&api.JobResponse{}, nil)
// 	
// 		err := s.consumerHandler.ProcessMsg(context.Background(), msg, jobMsg)
// 		require.NoError(t, err)
// 	
// 		select {
// 		case <-time.Tick(time.Millisecond * 500):
// 			t.Error(errMsgExceedTime)
// 		case rjob := <-cjob:
// 			assert.Equal(t, jobMsg.UUID, rjob.UUID)
// 		}
// 	})
// }
