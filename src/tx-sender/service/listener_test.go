// +build unit
// +build !race

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/encoding/proto"
	"github.com/consensys/orchestrate/pkg/errors"
	mock3 "github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	"github.com/consensys/orchestrate/pkg/types/tx"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/broker/sarama/mock"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/mocks"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const errMsgExceedTime = "exceeded waiting time"

type messageListenerCtrlTestSuite struct {
	suite.Suite
	listener            *MessageListener
	sendETHUC           *mocks.MockSendETHTxUseCase
	sendETHRawUC        *mocks.MockSendETHRawTxUseCase
	sendEEAPrivateUC    *mocks.MockSendEEAPrivateTxUseCase
	sendGoQuorumMarking *mocks.MockSendGoQuorumMarkingTxUseCase
	sendGoQuorumPrivate *mocks.MockSendGoQuorumPrivateTxUseCase
	apiClient           *mock3.MockOrchestrateClient
	tenantID            string
	allowedTenants      []string
	senderTopic         string
	recoverTopic        string
	consumerGroup       string
}

var _ usecases.UseCases = &messageListenerCtrlTestSuite{}

func (s *messageListenerCtrlTestSuite) SendETHRawTx() usecases.SendETHRawTxUseCase {
	return s.sendETHRawUC
}

func (s *messageListenerCtrlTestSuite) SendETHTx() usecases.SendETHTxUseCase {
	return s.sendETHUC
}

func (s *messageListenerCtrlTestSuite) SendEEAPrivateTx() usecases.SendEEAPrivateTxUseCase {
	return s.sendEEAPrivateUC
}

func (s *messageListenerCtrlTestSuite) SendGoQuorumPrivateTx() usecases.SendGoQuorumPrivateTxUseCase {
	return s.sendGoQuorumPrivate
}

func (s *messageListenerCtrlTestSuite) SendGoQuorumMarkingTx() usecases.SendGoQuorumMarkingTxUseCase {
	return s.sendGoQuorumMarking
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
	s.sendETHRawUC = mocks.NewMockSendETHRawTxUseCase(ctrl)
	s.sendETHUC = mocks.NewMockSendETHTxUseCase(ctrl)
	s.sendEEAPrivateUC = mocks.NewMockSendEEAPrivateTxUseCase(ctrl)
	s.sendGoQuorumPrivate = mocks.NewMockSendGoQuorumPrivateTxUseCase(ctrl)
	s.sendGoQuorumMarking = mocks.NewMockSendGoQuorumMarkingTxUseCase(ctrl)
	s.apiClient = mock3.NewMockOrchestrateClient(ctrl)
	s.senderTopic = "sender-topic"
	s.recoverTopic = "recover-topic"
	s.consumerGroup = "kafka-consumer-group"

	bckoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Millisecond*100), 2)
	s.listener = NewMessageListener(s, s.apiClient, s.recoverTopic, s.senderTopic, bckoff)
}

func (s *messageListenerCtrlTestSuite) TestMessageListener_PublicEthereum() {
	var claims map[string][]int32
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	session := mock.NewConsumerGroupSession(ctx, s.consumerGroup, claims)
	consumerClaim := mock.NewConsumerGroupClaim(s.senderTopic, 0, 0)

	defer func() {
		_ = s.listener.Cleanup(session)
	}()

	go func() {
		_ = s.listener.ConsumeClaim(session, consumerClaim)
	}()

	s.T().Run("should execute use case for multiple public ethereum transactions", func(t *testing.T) {
		envelope := fakeEnvelope(s.tenantID)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(envelope.TxEnvelopeAsRequest())
		
		cjob := make(chan *entities.Job, 1)
		s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
			cjob <- job
			return nil
		})

		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			assert.Equal(t, rjob.UUID, envelope.GetJobUUID())
		}
	})

	s.T().Run("should execute use case for public raw ethereum transactions", func(t *testing.T) {
		envelope := fakeEnvelope(s.tenantID)
		_ = envelope.SetJobType(tx.JobType_ETH_RAW_TX)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(envelope.TxEnvelopeAsRequest())

		cjob := make(chan *entities.Job, 1)
		s.sendETHRawUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
			cjob <- job
			return nil
		})

		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			assert.Equal(t, rjob.UUID, envelope.GetJobUUID())
		}
	})

	s.T().Run("should execute use case for eea transactions", func(t *testing.T) {
		envelope := fakeEnvelope(s.tenantID)
		_ = envelope.SetJobType(tx.JobType_EEA_PRIVATE_TX)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(envelope.TxEnvelopeAsRequest())

		cjob := make(chan *entities.Job, 1)
		s.sendEEAPrivateUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
			cjob <- job
			return nil
		})

		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			assert.Equal(t, rjob.UUID, envelope.GetJobUUID())
		}
	})

	s.T().Run("should execute use case for tessera marking transactions", func(t *testing.T) {
		envelope := fakeEnvelope(s.tenantID)
		_ = envelope.SetJobType(tx.JobType_GO_QUORUM_MARKING_TX)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(envelope.TxEnvelopeAsRequest())

		cjob := make(chan *entities.Job, 1)
		s.sendGoQuorumMarking.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
			cjob <- job
			return nil
		})
		
		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			assert.Equal(t, rjob.UUID, envelope.GetJobUUID())
		}
	})

	s.T().Run("should execute use case for tessera private transactions", func(t *testing.T) {
		envelope := fakeEnvelope(s.tenantID)
		_ = envelope.SetJobType(tx.JobType_GO_QUORUM_PRIVATE_TX)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(envelope.TxEnvelopeAsRequest())

		cjob := make(chan *entities.Job, 1)
		s.sendGoQuorumPrivate.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
			cjob <- job
			return nil
		})

		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			assert.Equal(t, rjob.UUID, envelope.GetJobUUID())
		}
	})
}

func (s *messageListenerCtrlTestSuite) TestMessageListener_PublicEthereum_Errors() {
	var claims map[string][]int32
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	session := mock.NewConsumerGroupSession(ctx, s.consumerGroup, claims)
	consumerClaim := mock.NewConsumerGroupClaim(s.senderTopic, 0, 0)

	defer func() {
		_ = s.listener.Cleanup(session)
	}()

	go func() {
		_ = s.listener.ConsumeClaim(session, consumerClaim)
	}()

	s.T().Run("should update transaction and send message to tx-recover if sending fails", func(t *testing.T) {
		expectedErr := errors.InternalError("error")
		evlp := fakeEnvelope(s.tenantID)
		_ = evlp.SetJobType(tx.JobType_ETH_TX)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(evlp.TxEnvelopeAsRequest())

		cjob := make(chan *entities.Job, 1)
		s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
			cjob <- job
			return expectedErr
		})
		s.apiClient.EXPECT().UpdateJob(gomock.Any(), evlp.GetJobUUID(), &api.UpdateJobRequest{
			Status:      entities.StatusFailed,
			Message:     expectedErr.Error(),
			Transaction: nil,
		}).Return(&api.JobResponse{}, nil)

		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			time.Sleep(time.Millisecond * 500) // Wait for receipt to be sent
			assert.Equal(t, evlp.GetJobUUID(), rjob.UUID)
		}
	})

	s.T().Run("should update transaction and retry job if sending fails by nonce error", func(t *testing.T) {
		invalidNonceErr := errors.InvalidNonceWarning("nonce too low")
		evlp := fakeEnvelope(s.tenantID)
		msg := &sarama.ConsumerMessage{}
		msg.Value, _ = proto.Marshal(evlp.TxEnvelopeAsRequest())

		cjob := make(chan *entities.Job, 1)
		gomock.InOrder(
			s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(invalidNonceErr),
			s.sendETHUC.EXPECT().Execute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job *entities.Job) error {
				cjob <- job
				return nil
			}),
		)

		s.apiClient.EXPECT().UpdateJob(gomock.Any(), evlp.GetJobUUID(), &api.UpdateJobRequest{
			Status:      entities.StatusRecovering,
			Message:     invalidNonceErr.Error(),
			Transaction: nil,
		}).Return(&api.JobResponse{}, nil)


		consumerClaim.ExpectMessage(msg)

		select {
		case <-ctx.Done():
			t.Error(ctx.Err())
		case <-time.Tick(time.Millisecond * 500):
			t.Error(errMsgExceedTime)
		case rjob := <-cjob:
			assert.Equal(t, evlp.GetJobUUID(), rjob.UUID)
		}
	})
}

func fakeEnvelope(tenantID string) *tx.Envelope {
	jobUUID := uuid.Must(uuid.NewV4()).String()
	scheduleUUID := uuid.Must(uuid.NewV4()).String()

	envelope := tx.NewEnvelope()
	_ = envelope.SetID(jobUUID)
	_ = envelope.SetJobUUID(jobUUID)
	_ = envelope.SetScheduleUUID(scheduleUUID)
	_ = envelope.SetNonce(0)
	_ = envelope.SetFromString("0xeca84382E0f1dDdE22EedCd0D803442972EC7BE5")
	_ = envelope.SetGas(21000)
	_ = envelope.SetGasPriceString("10000000")
	_ = envelope.SetValueString("10000000")
	_ = envelope.SetDataString("0x")
	_ = envelope.SetChainIDString("1")
	_ = envelope.SetHeadersValue(utils.TenantIDHeader, tenantID)
	_ = envelope.SetPrivateFrom("A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=")
	_ = envelope.SetPrivateFor([]string{"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=", "B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="})

	return envelope
}
