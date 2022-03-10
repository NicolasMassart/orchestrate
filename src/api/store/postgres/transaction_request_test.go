// +build unit

package postgres

import (
	"context"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/consensys/orchestrate/src/infra/postgres/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type pgTransactionRequestTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGTransactionRequest
}

func TestPGTransactionRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgTransactionRequestTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGTransactionRequest(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgTransactionRequestTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	requestHash := "requestHash"
	scheduleUUID := "scheduleUUID"

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeTxRequest := testdata.FakeTxRequest()
		mockQuery := mocks.NewMockQuery(ctrl)

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, &models.Schedule{}).Return(mockQuery),
			mockQuery.EXPECT().Column("id").Return(mockQuery),
			mockQuery.EXPECT().Where("uuid = ?", scheduleUUID).Return(mockQuery),
			mockQuery.EXPECT().Select().Return(nil),
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
				txRequestModel := mdls[0].(*models.TransactionRequest)
				assert.Equal(t, requestHash, txRequestModel.RequestHash)
				assert.Equal(t, fakeTxRequest.Params, txRequestModel.Params)
				assert.Equal(t, fakeTxRequest.ChainName, txRequestModel.ChainName)
				assert.Equal(t, fakeTxRequest.IdempotencyKey, txRequestModel.IdempotencyKey)

				return mockQuery
			}),
			mockQuery.EXPECT().Insert().Return(nil),
		)

		txRequest, err := s.dataAgent.Insert(ctx, fakeTxRequest, requestHash, scheduleUUID)
		require.NoError(t, err)

		assert.False(t, txRequest.CreatedAt.IsZero())
		assert.False(t, txRequest.Params.CreatedAt.IsZero())
		assert.False(t, txRequest.Params.UpdatedAt.IsZero())
		assert.Equal(t, requestHash, txRequest.Hash)
		assert.Equal(t, fakeTxRequest.IdempotencyKey, txRequest.IdempotencyKey)
		assert.Equal(t, fakeTxRequest.ChainName, txRequest.ChainName)
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, &models.Schedule{}).Return(mockQuery)
		mockQuery.EXPECT().Column("id").Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", scheduleUUID).Return(mockQuery)
		mockQuery.EXPECT().Select().Return(expectedErr)

		account, err := s.dataAgent.Insert(ctx, testdata.FakeTxRequest(), requestHash, scheduleUUID)
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, &models.Schedule{}).Return(mockQuery)
		mockQuery.EXPECT().Column("id").Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", scheduleUUID).Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)
		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(expectedErr)

		account, err := s.dataAgent.Insert(ctx, testdata.FakeTxRequest(), requestHash, scheduleUUID)
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgTransactionRequestTestSuite) TestFindOneByUUID() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	scheduleUUID := "scheduleUUID"
	tenants := []string{"tenant1", "tenant2"}
	owner := "owner"

	s.T().Run("should find successfully", func(t *testing.T) {
		fakeTxRequest := testdata.FakeTxRequest()
		mockQuery := mocks.NewMockQuery(ctrl)
		var txRequestModel *models.TransactionRequest

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, &models.TransactionRequest{}).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
				txRequestModel = mdls[0].(*models.TransactionRequest)
				return mockQuery
			}),
			mockQuery.EXPECT().Where("schedule.uuid = ?", scheduleUUID).Return(mockQuery),
			mockQuery.EXPECT().Relation("Schedule").Return(mockQuery),
			mockQuery.EXPECT().WhereAllowedTenants("schedule.tenant_id", tenants).Return(mockQuery),
			mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery),
			mockQuery.EXPECT().SelectOne().DoAndReturn(func() error {
				txRequestModel.RequestHash = fakeTxRequest.Hash
				txRequestModel.IdempotencyKey = fakeTxRequest.IdempotencyKey
				txRequestModel.ChainName = fakeTxRequest.ChainName
				txRequestModel.CreatedAt = fakeTxRequest.CreatedAt
				txRequestModel.Params = fakeTxRequest.Params

				return nil
			}),
		)

		txRequest, err := s.dataAgent.FindOneByUUID(ctx, scheduleUUID, tenants, owner)
		require.NoError(t, err)

		assert.Equal(t, fakeTxRequest.Hash, txRequest.Hash)
		assert.Equal(t, fakeTxRequest.IdempotencyKey, txRequest.IdempotencyKey)
		assert.Equal(t, fakeTxRequest.ChainName, txRequest.ChainName)
		assert.Equal(t, fakeTxRequest.CreatedAt, txRequest.CreatedAt)
		assert.Equal(t, fakeTxRequest.Params.CreatedAt, txRequest.Params.CreatedAt)
		assert.Equal(t, fakeTxRequest.Params.UpdatedAt, txRequest.Params.UpdatedAt)
		assert.Equal(t, fakeTxRequest.Params.UUID, txRequest.Params.UUID)
		assert.Equal(t, fakeTxRequest.Params.Hash, txRequest.Params.Hash)
		assert.Equal(t, fakeTxRequest.Params.To, txRequest.Params.To)
		assert.Equal(t, fakeTxRequest.Params.From, txRequest.Params.From)
		assert.Equal(t, fakeTxRequest.Params.TransactionType, txRequest.Params.TransactionType)
		assert.Equal(t, fakeTxRequest.Params.MandatoryFor, txRequest.Params.MandatoryFor)
		assert.Equal(t, fakeTxRequest.Params.MethodSignature, txRequest.Params.MethodSignature)
		assert.Equal(t, fakeTxRequest.Params.PrivateFrom, txRequest.Params.PrivateFrom)
		assert.Equal(t, fakeTxRequest.Params.PrivateFor, txRequest.Params.PrivateFor)
		assert.Equal(t, fakeTxRequest.Params.PrivacyGroupID, txRequest.Params.PrivacyGroupID)
		assert.Equal(t, fakeTxRequest.Params.GasTipCap, txRequest.Params.GasTipCap)
		assert.Equal(t, fakeTxRequest.Params.GasFeeCap, txRequest.Params.GasFeeCap)
		assert.Equal(t, fakeTxRequest.Params.AccessList, txRequest.Params.AccessList)
		assert.Equal(t, fakeTxRequest.Params.Nonce, txRequest.Params.Nonce)
		assert.Equal(t, fakeTxRequest.Params.PrivacyFlag, txRequest.Params.PrivacyFlag)
		assert.Equal(t, fakeTxRequest.Params.Protocol, txRequest.Params.Protocol)
		assert.Equal(t, fakeTxRequest.Params.GasPrice, txRequest.Params.GasPrice)
		assert.Equal(t, fakeTxRequest.Params.ContractName, txRequest.Params.ContractName)
		assert.Equal(t, fakeTxRequest.Params.ContractTag, txRequest.Params.ContractTag)
	})

	s.T().Run("should fail with same error if SelectOne fails", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		expectedErr := errors.PostgresConnectionError("error")

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, &models.TransactionRequest{}).Return(mockQuery),
			mockQuery.EXPECT().Where("schedule.uuid = ?", scheduleUUID).Return(mockQuery),
			mockQuery.EXPECT().Relation("Schedule").Return(mockQuery),
			mockQuery.EXPECT().WhereAllowedTenants("schedule.tenant_id", tenants).Return(mockQuery),
			mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery),
			mockQuery.EXPECT().SelectOne().Return(expectedErr),
		)

		txRequest, err := s.dataAgent.FindOneByUUID(ctx, scheduleUUID, tenants, owner)
		assert.Nil(t, txRequest)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgTransactionRequestTestSuite) TestFindOneByIdempotencyKey() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	idempotencyKey := "idempotencyKey"
	tenantID := "tenant1"
	owner := "owner"

	s.T().Run("should find successfully", func(t *testing.T) {
		fakeTxRequest := testdata.FakeTxRequest()
		mockQuery := mocks.NewMockQuery(ctrl)
		var txRequestModel *models.TransactionRequest

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, &models.TransactionRequest{}).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
				txRequestModel = mdls[0].(*models.TransactionRequest)
				return mockQuery
			}),
			mockQuery.EXPECT().Where("idempotency_key = ?", idempotencyKey).Return(mockQuery),
			mockQuery.EXPECT().Relation("Schedule").Return(mockQuery),
			mockQuery.EXPECT().Where("schedule.tenant_id = ?", tenantID).Return(mockQuery),
			mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery),
			mockQuery.EXPECT().SelectOne().DoAndReturn(func() error {
				txRequestModel.RequestHash = fakeTxRequest.Hash
				txRequestModel.IdempotencyKey = fakeTxRequest.IdempotencyKey
				txRequestModel.ChainName = fakeTxRequest.ChainName
				txRequestModel.CreatedAt = fakeTxRequest.CreatedAt
				txRequestModel.Params = fakeTxRequest.Params

				return nil
			}),
		)

		txRequest, err := s.dataAgent.FindOneByIdempotencyKey(ctx, idempotencyKey, tenantID, owner)
		require.NoError(t, err)

		assert.Equal(t, fakeTxRequest.Hash, txRequest.Hash)
		assert.Equal(t, fakeTxRequest.IdempotencyKey, txRequest.IdempotencyKey)
		assert.Equal(t, fakeTxRequest.ChainName, txRequest.ChainName)
		assert.Equal(t, fakeTxRequest.CreatedAt, txRequest.CreatedAt)
		assert.Equal(t, fakeTxRequest.Params.CreatedAt, txRequest.Params.CreatedAt)
		assert.Equal(t, fakeTxRequest.Params.UpdatedAt, txRequest.Params.UpdatedAt)
		assert.Equal(t, fakeTxRequest.Params.UUID, txRequest.Params.UUID)
		assert.Equal(t, fakeTxRequest.Params.Hash, txRequest.Params.Hash)
		assert.Equal(t, fakeTxRequest.Params.To, txRequest.Params.To)
		assert.Equal(t, fakeTxRequest.Params.From, txRequest.Params.From)
		assert.Equal(t, fakeTxRequest.Params.TransactionType, txRequest.Params.TransactionType)
		assert.Equal(t, fakeTxRequest.Params.MandatoryFor, txRequest.Params.MandatoryFor)
		assert.Equal(t, fakeTxRequest.Params.MethodSignature, txRequest.Params.MethodSignature)
		assert.Equal(t, fakeTxRequest.Params.PrivateFrom, txRequest.Params.PrivateFrom)
		assert.Equal(t, fakeTxRequest.Params.PrivateFor, txRequest.Params.PrivateFor)
		assert.Equal(t, fakeTxRequest.Params.PrivacyGroupID, txRequest.Params.PrivacyGroupID)
		assert.Equal(t, fakeTxRequest.Params.GasTipCap, txRequest.Params.GasTipCap)
		assert.Equal(t, fakeTxRequest.Params.GasFeeCap, txRequest.Params.GasFeeCap)
		assert.Equal(t, fakeTxRequest.Params.AccessList, txRequest.Params.AccessList)
		assert.Equal(t, fakeTxRequest.Params.Nonce, txRequest.Params.Nonce)
		assert.Equal(t, fakeTxRequest.Params.PrivacyFlag, txRequest.Params.PrivacyFlag)
		assert.Equal(t, fakeTxRequest.Params.Protocol, txRequest.Params.Protocol)
		assert.Equal(t, fakeTxRequest.Params.GasPrice, txRequest.Params.GasPrice)
		assert.Equal(t, fakeTxRequest.Params.ContractName, txRequest.Params.ContractName)
		assert.Equal(t, fakeTxRequest.Params.ContractTag, txRequest.Params.ContractTag)
	})

	s.T().Run("should fail with same error if SelectOne fails", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		expectedErr := errors.PostgresConnectionError("error")

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, &models.TransactionRequest{}).Return(mockQuery),
			mockQuery.EXPECT().Where("idempotency_key = ?", idempotencyKey).Return(mockQuery),
			mockQuery.EXPECT().Relation("Schedule").Return(mockQuery),
			mockQuery.EXPECT().Where("schedule.tenant_id = ?", tenantID).Return(mockQuery),
			mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery),
			mockQuery.EXPECT().SelectOne().Return(expectedErr),
		)

		txRequest, err := s.dataAgent.FindOneByIdempotencyKey(ctx, idempotencyKey, tenantID, owner)
		assert.Nil(t, txRequest)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgTransactionRequestTestSuite) TestSearch() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}
	owner := "owner"

	s.T().Run("should search successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Relation("Schedule").Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("schedule.tenant_id", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)

		_, err := s.dataAgent.Search(ctx, &entities.TransactionRequestFilters{}, tenants, owner)
		require.NoError(t, err)
	})

	s.T().Run("should search successfully with filters", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		filters := &entities.TransactionRequestFilters{
			IdempotencyKeys: []string{"idempotencyKey1", "idempotencyKey2"},
		}

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Relation("Schedule").Return(mockQuery)
		mockQuery.EXPECT().Where("transaction_request.idempotency_key in (?)", gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("schedule.tenant_id", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)

		_, err := s.dataAgent.Search(ctx, filters, tenants, owner)
		require.NoError(t, err)
	})

	s.T().Run("should return empty array if Select fails with NotFoundError", func(t *testing.T) {
		notFoundErr := errors.NotFoundError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Relation("Schedule").Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("schedule.tenant_id", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery)
		mockQuery.EXPECT().Select().Return(notFoundErr)

		txRequests, err := s.dataAgent.Search(ctx, &entities.TransactionRequestFilters{}, tenants, owner)
		require.NoError(t, err)

		assert.Empty(t, txRequests)
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Relation("Schedule").Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("schedule.tenant_id", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("schedule.owner_id", owner).Return(mockQuery)
		mockQuery.EXPECT().Select().Return(expectedErr)

		txRequests, err := s.dataAgent.Search(ctx, &entities.TransactionRequestFilters{}, tenants, owner)
		assert.Empty(t, txRequests)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}
