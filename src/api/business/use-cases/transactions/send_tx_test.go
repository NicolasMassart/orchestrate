// +build unit

package transactions

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	mocks2 "github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type sendTxSuite struct {
	suite.Suite
	usecase            usecases.SendTxUseCase
	DB                 *mocks2.MockDB
	SearchChainsUC     *mocks.MockSearchChainsUseCase
	TxRequestDA        *mocks2.MockTransactionRequestAgent
	ScheduleDA         *mocks2.MockScheduleAgent
	StartJobUC         *mocks.MockStartJobUseCase
	CreateJobUC        *mocks.MockCreateJobUseCase
	GetTxUC            *mocks.MockGetTxUseCase
	GetFaucetCandidate *mocks.MockGetFaucetCandidateUseCase
	userInfo           *multitenancy.UserInfo
}

var (
	faucetNotFoundErr = errors.NotFoundError("not found faucet candidate")
)

func TestSendTx(t *testing.T) {
	s := new(sendTxSuite)
	suite.Run(t, s)
}

func (s *sendTxSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.DB = mocks2.NewMockDB(ctrl)
	s.SearchChainsUC = mocks.NewMockSearchChainsUseCase(ctrl)
	s.TxRequestDA = mocks2.NewMockTransactionRequestAgent(ctrl)
	s.ScheduleDA = mocks2.NewMockScheduleAgent(ctrl)
	s.StartJobUC = mocks.NewMockStartJobUseCase(ctrl)
	s.CreateJobUC = mocks.NewMockCreateJobUseCase(ctrl)
	s.GetTxUC = mocks.NewMockGetTxUseCase(ctrl)
	s.GetFaucetCandidate = mocks.NewMockGetFaucetCandidateUseCase(ctrl)

	s.DB.EXPECT().TransactionRequest().Return(s.TxRequestDA).AnyTimes()
	s.DB.EXPECT().Schedule().Return(s.ScheduleDA).AnyTimes()
	s.DB.EXPECT().Schedule().Return(s.ScheduleDA).AnyTimes()
	s.DB.EXPECT().TransactionRequest().Return(s.TxRequestDA).AnyTimes()
	s.userInfo = multitenancy.NewUserInfo("tenantOne", "username")

	s.usecase = NewSendTxUseCase(
		s.DB,
		s.SearchChainsUC,
		s.StartJobUC,
		s.CreateJobUC,
		s.GetTxUC,
		s.GetFaucetCandidate,
	)
}

func (s *sendTxSuite) TestSendTx_Success() {
	s.T().Run("should execute send successfully a public tx", func(t *testing.T) {
		txRequest := testdata.FakeTxRequest()

		response, err := successfulTestExecution(s, txRequest, false, entities.EthereumTransaction)
		assert.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should execute send successfully a public tx with faucet", func(t *testing.T) {
		txRequest := testdata.FakeTxRequest()

		response, err := successfulTestExecution(s, txRequest, true, entities.EthereumTransaction)
		assert.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should execute send successfully a EEA tx", func(t *testing.T) {
		txRequest := testdata.FakeEEATxRequest()
		txRequest.Params.Protocol = entities.EEAChainType

		response, err := successfulTestExecution(s, txRequest, false, entities.EEAPrivateTransaction,
			entities.EEAMarkingTransaction)
		assert.NoError(t, err)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should execute send successfully a tessera tx", func(t *testing.T) {
		txRequest := testdata.FakeTesseraTxRequest()
		txRequest.Params.Protocol = entities.GoQuorumChainType
	
		response, err := successfulTestExecution(s, txRequest, false, entities.TesseraPrivateTransaction,
			entities.TesseraMarkingTransaction)
		assert.NoError(t, err)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should execute send successfully a raw tx", func(t *testing.T) {
		txRequest := testdata.FakeRawTxRequest()
		txRequest.Params.Raw = hexutil.MustDecode("0xf85380839896808252088083989680808216b4a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1e")

		response, err := successfulTestExecution(s, txRequest, false, entities.EthereumRawTransaction)
		assert.NoError(t, err)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should not insert and start job in Postgres if TxRequest already exists and send if status is CREATED", func(t *testing.T) {
		txRequest := testdata.FakeTxRequest()
		ctx := context.Background()
		chains := []*entities.Chain{testdata.FakeChain()}
		txRequest.Hash, _ = generateRequestHash(chains[0].UUID, txRequest.Params)
		jobUUID := txRequest.Schedule.Jobs[0].UUID
		txData := (hexutil.Bytes)(hexutil.MustDecode("0x"))

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(txRequest, nil)
		s.GetTxUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, s.userInfo).Return(txRequest, nil)
		s.GetFaucetCandidate.EXPECT().Execute(gomock.Any(), gomock.Any(), chains[0], s.userInfo).
			Return(nil, faucetNotFoundErr)
		s.StartJobUC.EXPECT().Execute(gomock.Any(), jobUUID, s.userInfo).Return(nil)
		s.GetTxUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, s.userInfo).Return(txRequest, nil)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)

		require.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should not insert and not start job if TxRequest already exists and not send if status is not CREATED", func(t *testing.T) {
		txRequest := testdata.FakeTxRequest()
		txRequest.Schedule.Jobs[0].Status = entities.StatusStarted

		ctx := context.Background()
		txData := (hexutil.Bytes)(hexutil.MustDecode("0x"))
		chains := []*entities.Chain{testdata.FakeChain()}
		txRequest.Hash, _ = generateRequestHash(chains[0].UUID, txRequest.Params)

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(txRequest, nil)
		s.GetTxUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, s.userInfo).Return(txRequest, nil)
		s.GetTxUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, s.userInfo).Return(txRequest, nil)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)
		require.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
	})

	s.T().Run("should execute send successfully a oneTimeKey tx", func(t *testing.T) {
		txRequest := testdata.FakeTxRequest()
		txRequest.Params.From = nil
		txRequest.InternalData.OneTimeKey = true

		response, err := successfulTestExecution(s, txRequest, false, entities.EthereumTransaction)
		assert.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
		assert.Equal(t, txRequest.IdempotencyKey, response.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, response.Schedule.UUID)
		assert.True(t, response.Schedule.Jobs[0].InternalData.OneTimeKey)
	})
}

func (s *sendTxSuite) TestSendTx_ExpectedErrors() {
	ctx := context.Background()
	chains := []*entities.Chain{testdata.FakeChain()}
	txData := (hexutil.Bytes)(hexutil.MustDecode("0x"))

	s.T().Run("should fail with same error if chain agent fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")
		txRequest := testdata.FakeTxRequest()

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).
			Return(nil, expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)
		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with InvalidParameterError if no chain is found", func(t *testing.T) {
		txRequest := testdata.FakeTxRequest()

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).
			Return([]*entities.Chain{}, nil)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)
		assert.Nil(t, response)
		assert.True(t, errors.IsInvalidParameterError(err))
	})

	s.T().Run("should fail with same error if FindOne fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		txRequest := testdata.FakeTxRequest()

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(nil, expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)
		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with AlreadyExistsError if request found has different request hash", func(t *testing.T) {
		expectedErr := errors.AlreadyExistsError("transaction request with the same idempotency key and different params already exists")
		txRequest := testdata.FakeTxRequest()

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(&entities.TxRequest{
			Hash: "differentRequestHash",
		}, nil)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)
		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with same error if RunInTransaction fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		txRequest := testdata.FakeTxRequest()

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).
			Return(nil, errors.NotFoundError(""))
		s.DB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with same error if createJob UseCase fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		txRequest := testdata.FakeTxRequest()
		txData := (hexutil.Bytes)(hexutil.MustDecode("0x"))
		txRequest.Schedule = testdata.FakeSchedule()
		txRequest.Schedule.TenantID = s.userInfo.TenantID
		txRequest.Schedule.OwnerID = s.userInfo.Username

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}}, s.userInfo).
			Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(nil, errors.NotFoundError(""))
		s.ScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), txRequest.Schedule.UUID, s.userInfo.AllowedTenants, s.userInfo.Username).
			Return(txRequest.Schedule, nil)
		s.DB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)
		s.GetFaucetCandidate.EXPECT().Execute(gomock.Any(), *txRequest.Schedule.Jobs[0].Transaction.From, chains[0], s.userInfo).
			Return(nil, faucetNotFoundErr)
		s.CreateJobUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).
			Return(txRequest.Schedule.Jobs[0], expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with same error if getFaucetCandidate request fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		chains := []*entities.Chain{testdata.FakeChain()}
		txRequest := testdata.FakeTxRequest()
		txData := (hexutil.Bytes)(hexutil.MustDecode("0x"))
		txRequest.Schedule.TenantID = s.userInfo.TenantID
		txRequest.Schedule.OwnerID = s.userInfo.Username

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}}, s.userInfo).
			Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(nil, errors.NotFoundError(""))
		s.DB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)
		s.ScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), txRequest.Schedule.UUID, s.userInfo.AllowedTenants, s.userInfo.Username).
			Return(txRequest.Schedule, nil)
		s.CreateJobUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(txRequest.Schedule.Jobs[0], nil)
		s.GetFaucetCandidate.EXPECT().Execute(gomock.Any(), *txRequest.Schedule.Jobs[0].Transaction.From, gomock.Any(), s.userInfo).
			Return(nil, expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with same error if startJob UseCase fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		txRequest := testdata.FakeTxRequest()
		txRequest.Schedule = testdata.FakeSchedule()
		txRequest.Schedule.TenantID = s.userInfo.TenantID
		txRequest.Schedule.OwnerID = s.userInfo.Username

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(nil, errors.NotFoundError(""))
		s.DB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)
		s.ScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), txRequest.Schedule.UUID, s.userInfo.AllowedTenants, s.userInfo.Username).
			Return(txRequest.Schedule, nil)
		s.CreateJobUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(txRequest.Schedule.Jobs[0], nil)
		s.GetFaucetCandidate.EXPECT().Execute(gomock.Any(), *txRequest.Schedule.Jobs[0].Transaction.From, gomock.Any(), s.userInfo).Return(nil, faucetNotFoundErr)
		s.StartJobUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.Jobs[0].UUID, s.userInfo).Return(expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})

	s.T().Run("should fail with same error if getTx UseCase fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		txRequest := testdata.FakeTxRequest()

		s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
			s.userInfo).Return(chains, nil)

		requestHash, _ := generateRequestHash(chains[0].UUID, txRequest.Params)
		s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
			s.userInfo.Username).Return(&entities.TxRequest{
			Hash:     requestHash,
			Schedule: txRequest.Schedule,
		}, nil)
		s.GetTxUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, s.userInfo).Return(nil, expectedErr)

		response, err := s.usecase.Execute(ctx, txRequest, txData, s.userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(sendTxComponent), err)
	})
}

func successfulTestExecution(s *sendTxSuite, txRequest *entities.TxRequest, withFaucet bool, jobTypes ...entities.JobType) (*entities.TxRequest, error) {
	ctx := context.Background()
	chains := []*entities.Chain{testdata.FakeChain()}
	jobUUID := txRequest.Schedule.Jobs[0].UUID
	txData := (hexutil.Bytes)(hexutil.MustDecode("0x"))
	txRequest.Schedule.TenantID = s.userInfo.TenantID
	txRequest.Schedule.OwnerID = s.userInfo.Username
	jobIdx := 0
	from := new(ethcommon.Address)
	if txRequest.Params.From != nil {
		from = txRequest.Params.From
	}

	s.SearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{txRequest.ChainName}},
		s.userInfo).Return(chains, nil)
	s.TxRequestDA.EXPECT().FindOneByIdempotencyKey(gomock.Any(), txRequest.IdempotencyKey, s.userInfo.TenantID,
		s.userInfo.Username).Return(nil, errors.NotFoundError(""))
	s.ScheduleDA.EXPECT().FindOneByUUID(gomock.Any(), txRequest.Schedule.UUID, s.userInfo.AllowedTenants, s.userInfo.Username).
		Return(txRequest.Schedule, nil)
	s.DB.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(nil)
	s.CreateJobUC.EXPECT().
		Execute(gomock.Any(), gomock.Any(), s.userInfo).
		DoAndReturn(func(ctx context.Context, jobEntity *entities.Job, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
			if jobEntity.Type != jobTypes[jobIdx] {
				return nil, fmt.Errorf("invalid job type. Got %s, expected %s", jobEntity.Type, jobTypes[jobIdx])
			}

			jobEntity.Transaction.From = from
			jobEntity.UUID = jobUUID
			jobEntity.Status = entities.StatusCreated
			jobIdx = jobIdx + 1
			return jobEntity, nil
		}).Times(len(jobTypes))

	// We flag this "special" scenario as faucet funding tx flow
	if withFaucet {
		internalAdminUser := multitenancy.NewInternalAdminUser()
		internalAdminUser.TenantID = s.userInfo.TenantID
		faucet := testdata.FakeFaucet()
		s.GetFaucetCandidate.EXPECT().Execute(gomock.Any(), *from, chains[0], s.userInfo).Return(faucet, nil)
		s.CreateJobUC.EXPECT().Execute(gomock.Any(), gomock.Any(), internalAdminUser).
			DoAndReturn(func(ctx context.Context, jobEntity *entities.Job, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
				if jobEntity.Transaction.From.String() != faucet.CreditorAccount.String() {
					return nil, fmt.Errorf("invalid from account. Got %s, expected %s", jobEntity.Transaction.From, faucet.CreditorAccount)
				}

				jobEntity.UUID = faucet.UUID
				return jobEntity, nil
			})
		s.StartJobUC.EXPECT().Execute(gomock.Any(), faucet.UUID, internalAdminUser).Return(nil)
	} else {
		s.GetFaucetCandidate.EXPECT().Execute(gomock.Any(), *from, chains[0], s.userInfo).Return(nil, faucetNotFoundErr)
	}

	s.StartJobUC.EXPECT().Execute(gomock.Any(), jobUUID, s.userInfo).Return(nil)

	return s.usecase.Execute(ctx, txRequest, txData, s.userInfo)
}
