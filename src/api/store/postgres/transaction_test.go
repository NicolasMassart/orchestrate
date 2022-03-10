// +build unit

package postgres

import (
	"context"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/consensys/orchestrate/src/infra/postgres/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type pgTransactionTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGTransaction
}

func TestPGTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgTransactionTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGTransaction(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgTransactionTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeTx := testdata.FakeETHTransaction()
		fakeTx.UUID = ""
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			txModel := mdls[0].(*models.Transaction)
			assert.NotEmpty(t, txModel.UUID)
			assert.Equal(t, fakeTx.From.Hex(), txModel.Sender)
			assert.Equal(t, fakeTx.Value.ToInt().String(), txModel.Value)
			assert.Equal(t, fakeTx.AccessList, txModel.AccessList)
			assert.Equal(t, fakeTx.Data.String(), txModel.Data)
			//	assert.Equal(t, fakeTx.Gas, txModel.Gas)
			assert.Equal(t, fakeTx.GasPrice.ToInt().String(), txModel.GasPrice)
			//			assert.Equal(t, fakeTx.GasTipCap.ToInt().String(), txModel.GasTipCap)
			assert.Equal(t, fakeTx.Hash.Hex(), txModel.Hash)
			//			assert.Equal(t, fakeTx.Nonce, txModel.Nonce)
			//	assert.Equal(t, int(fakeTx.PrivacyFlag), txModel.PrivacyFlag)
			assert.Equal(t, fakeTx.PrivacyGroupID, txModel.PrivacyGroupID)
			assert.Equal(t, fakeTx.PrivateFor, txModel.PrivateFor)
			assert.Equal(t, fakeTx.PrivateFrom, txModel.PrivateFrom)
			assert.Equal(t, fakeTx.MandatoryFor, txModel.MandatoryFor)
			assert.Equal(t, fakeTx.Raw.String(), txModel.Raw)
			assert.Equal(t, fakeTx.To.Hex(), txModel.Recipient)
			assert.Equal(t, fakeTx.TransactionType.String(), txModel.TxType)

			return mockQuery
		})
		mockQuery.EXPECT().Insert().Return(nil)

		tx, err := s.dataAgent.Insert(ctx, fakeTx)
		require.NoError(t, err)

		assert.False(t, tx.CreatedAt.IsZero())
		assert.False(t, tx.UpdatedAt.IsZero())
		assert.NotEmpty(t, tx.UUID)
		assert.Equal(t, fakeTx.From, tx.From)
		assert.Equal(t, fakeTx.Value, tx.Value)
		assert.Equal(t, fakeTx.AccessList, tx.AccessList)
		assert.Equal(t, fakeTx.Data, tx.Data)
		assert.Equal(t, fakeTx.EnclaveKey, tx.EnclaveKey)
		assert.Equal(t, fakeTx.Gas, tx.Gas)
		assert.Equal(t, fakeTx.GasPrice, tx.GasPrice)
		assert.Equal(t, fakeTx.GasFeeCap, tx.GasFeeCap)
		assert.Equal(t, fakeTx.GasTipCap, tx.GasTipCap)
		assert.Equal(t, fakeTx.Hash, tx.Hash)
		assert.Equal(t, fakeTx.Nonce, tx.Nonce)
		assert.Equal(t, fakeTx.PrivacyFlag, tx.PrivacyFlag)
		assert.Equal(t, fakeTx.PrivacyGroupID, tx.PrivacyGroupID)
		assert.Equal(t, fakeTx.PrivateFor, tx.PrivateFor)
		assert.Equal(t, fakeTx.PrivateFrom, tx.PrivateFrom)
		assert.Equal(t, fakeTx.MandatoryFor, tx.MandatoryFor)
		assert.Equal(t, fakeTx.Raw, tx.Raw)
		assert.Equal(t, fakeTx.To, tx.To)
		assert.Equal(t, fakeTx.TransactionType, tx.TransactionType)

	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(expectedErr)

		account, err := s.dataAgent.Insert(ctx, testdata.FakeETHTransaction())
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgTransactionTestSuite) TestUpdate() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should update and parse successfully", func(t *testing.T) {
		fakeTx := testdata.FakeETHTransaction()
		fakeTx.UpdatedAt = time.Time{}
		fakeTx.CreatedAt = time.Time{}
		mockQuery := mocks.NewMockQuery(ctrl)
		jobUUID := "jobUUID"
		uuid := "uuid"

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
				mdls[0].(*models.Transaction).UUID = "uuid"
				return mockQuery
			}),
			mockQuery.EXPECT().Join("JOIN jobs AS j ON j.transaction_id = transaction.id", gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("j.uuid = ?", jobUUID).Return(mockQuery),
			mockQuery.EXPECT().SelectOne().Return(nil),
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery),
			mockQuery.EXPECT().Update().Return(nil),
		)

		tx, err := s.dataAgent.Update(ctx, fakeTx, jobUUID)
		require.NoError(t, err)

		assert.False(t, tx.UpdatedAt.Equal(tx.CreatedAt))
		assert.Equal(t, fakeTx.UUID, tx.UUID)
		assert.Equal(t, tx.From, tx.From)
		assert.Equal(t, tx.Value, tx.Value)
		assert.Equal(t, tx.AccessList, tx.AccessList)
		assert.Equal(t, tx.Data, tx.Data)
		assert.Equal(t, tx.EnclaveKey, tx.EnclaveKey)
		assert.Equal(t, tx.Gas, tx.Gas)
		assert.Equal(t, tx.GasPrice, tx.GasPrice)
		assert.Equal(t, tx.GasFeeCap, tx.GasFeeCap)
		assert.Equal(t, tx.GasTipCap, tx.GasTipCap)
		assert.Equal(t, tx.Hash, tx.Hash)
		assert.Equal(t, tx.Nonce, tx.Nonce)
		assert.Equal(t, tx.PrivacyFlag, tx.PrivacyFlag)
		assert.Equal(t, tx.PrivacyGroupID, tx.PrivacyGroupID)
		assert.Equal(t, tx.PrivateFor, tx.PrivateFor)
		assert.Equal(t, tx.PrivateFrom, tx.PrivateFrom)
		assert.Equal(t, tx.MandatoryFor, tx.MandatoryFor)
		assert.Equal(t, tx.Raw, tx.Raw)
		assert.Equal(t, tx.To, tx.To)
		assert.Equal(t, tx.TransactionType, tx.TransactionType)
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Join("JOIN jobs AS j ON j.transaction_id = transaction.id", gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("j.uuid = ?", "jobUUID").Return(mockQuery),
			mockQuery.EXPECT().SelectOne().Return(expectedErr),
		)

		account, err := s.dataAgent.Update(ctx, testdata.FakeETHTransaction(), "jobUUID")
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})

	s.T().Run("should fail with same error if Update fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Join("JOIN jobs AS j ON j.transaction_id = transaction.id", gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("j.uuid = ?", "jobUUID").Return(mockQuery),
			mockQuery.EXPECT().SelectOne().Return(nil),
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("uuid = ?", gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Update().Return(expectedErr),
		)

		account, err := s.dataAgent.Update(ctx, testdata.FakeETHTransaction(), "jobUUID")
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}
