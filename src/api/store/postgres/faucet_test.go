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
	"time"
)

type pgFaucetTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGFaucet
}

func TestPGFaucet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgFaucetTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGFaucet(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgFaucetTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeFaucet := testdata.FakeFaucet()
		fakeFaucet.UUID = ""
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			faucetModel := mdls[0].(*models.Faucet)
			assert.NotEmpty(t, faucetModel.UUID)
			assert.Equal(t, fakeFaucet.Name, faucetModel.Name)
			assert.Equal(t, fakeFaucet.ChainRule, faucetModel.ChainRule)
			assert.Equal(t, fakeFaucet.CreditorAccount.Hex(), faucetModel.CreditorAccount)
			assert.Equal(t, fakeFaucet.TenantID, faucetModel.TenantID)
			assert.Equal(t, fakeFaucet.MaxBalance.ToInt().String(), faucetModel.MaxBalance)
			assert.Equal(t, fakeFaucet.Amount.ToInt().String(), faucetModel.Amount)
			assert.Equal(t, fakeFaucet.Cooldown, faucetModel.Cooldown)

			return mockQuery
		})
		mockQuery.EXPECT().Insert().Return(nil)

		faucet, err := s.dataAgent.Insert(ctx, fakeFaucet)
		require.NoError(t, err)

		assert.False(t, faucet.CreatedAt.IsZero())
		assert.False(t, faucet.UpdatedAt.IsZero())
		assert.NotEmpty(t, faucet.UUID)
		assert.Equal(t, fakeFaucet.Name, faucet.Name)
		assert.Equal(t, fakeFaucet.ChainRule, faucet.ChainRule)
		assert.Equal(t, fakeFaucet.CreditorAccount, faucet.CreditorAccount)
		assert.Equal(t, fakeFaucet.TenantID, faucet.TenantID)
		assert.Equal(t, fakeFaucet.MaxBalance, faucet.MaxBalance)
		assert.Equal(t, fakeFaucet.Amount, faucet.Amount)
		assert.Equal(t, fakeFaucet.Cooldown, faucet.Cooldown)
	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(expectedErr)

		account, err := s.dataAgent.Insert(ctx, testdata.FakeFaucet())
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgFaucetTestSuite) TestUpdate() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}

	s.T().Run("should update and parse successfully", func(t *testing.T) {
		fakeFaucet := testdata.FakeFaucet()
		fakeFaucet.UpdatedAt = time.Time{}
		fakeFaucet.CreatedAt = time.Time{}
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			faucetModel := mdls[0].(*models.Faucet)
			assert.Equal(t, fakeFaucet.UUID, faucetModel.UUID)
			assert.Equal(t, fakeFaucet.Name, faucetModel.Name)
			assert.Equal(t, fakeFaucet.ChainRule, faucetModel.ChainRule)
			assert.Equal(t, fakeFaucet.CreditorAccount.Hex(), faucetModel.CreditorAccount)
			assert.Equal(t, fakeFaucet.TenantID, faucetModel.TenantID)
			assert.Equal(t, fakeFaucet.MaxBalance.ToInt().String(), faucetModel.MaxBalance)
			assert.Equal(t, fakeFaucet.Amount.ToInt().String(), faucetModel.Amount)
			assert.Equal(t, fakeFaucet.Cooldown, faucetModel.Cooldown)

			return mockQuery
		})
		mockQuery.EXPECT().Where("uuid = ?", fakeFaucet.UUID).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Update().Return(nil)

		faucet, err := s.dataAgent.Update(ctx, fakeFaucet, tenants)
		require.NoError(t, err)

		assert.False(t, faucet.UpdatedAt.Equal(faucet.CreatedAt))
		assert.Equal(t, fakeFaucet.UUID, faucet.UUID)
		assert.Equal(t, fakeFaucet.Name, faucet.Name)
		assert.Equal(t, fakeFaucet.ChainRule, faucet.ChainRule)
		assert.Equal(t, fakeFaucet.CreditorAccount, faucet.CreditorAccount)
		assert.Equal(t, fakeFaucet.TenantID, faucet.TenantID)
		assert.Equal(t, fakeFaucet.MaxBalance, faucet.MaxBalance)
		assert.Equal(t, fakeFaucet.Amount, faucet.Amount)
		assert.Equal(t, fakeFaucet.Cooldown, faucet.Cooldown)
	})

	s.T().Run("should fail with same error if Update fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Update().Return(expectedErr)

		account, err := s.dataAgent.Update(ctx, testdata.FakeFaucet(), tenants)
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgFaucetTestSuite) TestDelete() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}

	s.T().Run("should delete successfully", func(t *testing.T) {
		fakeFaucet := testdata.FakeFaucet()
		fakeFaucet.UpdatedAt = time.Time{}
		fakeFaucet.CreatedAt = time.Time{}
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", fakeFaucet.UUID).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Delete().Return(nil)

		err := s.dataAgent.Delete(ctx, fakeFaucet, tenants)
		require.NoError(t, err)
	})

	s.T().Run("should fail with same error if Delete fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Delete().Return(expectedErr)

		err := s.dataAgent.Delete(ctx, testdata.FakeFaucet(), tenants)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgFaucetTestSuite) TestSearch() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}

	s.T().Run("should search successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Order("created_at ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)

		_, err := s.dataAgent.Search(ctx, &entities.FaucetFilters{}, tenants)
		require.NoError(t, err)
	})

	s.T().Run("should search successfully with filters", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		filters := &entities.FaucetFilters{
			Names:     []string{"name1", "name2"},
			TenantID:  "tenant",
			ChainRule: "chainRule",
		}

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		gomock.InOrder(
			mockQuery.EXPECT().Where("name in (?)", gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("tenant_id = ?", filters.TenantID).Return(mockQuery),
			mockQuery.EXPECT().Where("chain_rule = ?", filters.ChainRule).Return(mockQuery),
		)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Order("created_at ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)

		_, err := s.dataAgent.Search(ctx, filters, tenants)
		require.NoError(t, err)
	})

	s.T().Run("should return empty array if Select fails with NotFoundError", func(t *testing.T) {
		notFoundErr := errors.NotFoundError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Order("created_at ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(notFoundErr)

		accounts, err := s.dataAgent.Search(ctx, &entities.FaucetFilters{}, tenants)
		require.NoError(t, err)

		assert.Empty(t, accounts)
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().Order("created_at ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(expectedErr)

		accounts, err := s.dataAgent.Search(ctx, &entities.FaucetFilters{}, tenants)
		assert.Empty(t, accounts)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgFaucetTestSuite) TestFindOneByUUID() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}
	uuid := "uuid"

	s.T().Run("should find one successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		expectedFaucet := testdata.FakeFaucet()

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			assert.Equal(t, &models.Faucet{}, mdls[0].(*models.Faucet))

			fakeModel := models.NewFaucet(expectedFaucet)
			mdls[0].(*models.Faucet).Amount = fakeModel.Amount
			mdls[0].(*models.Faucet).UUID = fakeModel.UUID
			mdls[0].(*models.Faucet).MaxBalance = fakeModel.MaxBalance
			mdls[0].(*models.Faucet).ChainRule = fakeModel.ChainRule
			mdls[0].(*models.Faucet).Cooldown = fakeModel.Cooldown
			mdls[0].(*models.Faucet).CreditorAccount = fakeModel.CreditorAccount
			mdls[0].(*models.Faucet).Name = fakeModel.Name
			mdls[0].(*models.Faucet).TenantID = fakeModel.TenantID
			mdls[0].(*models.Faucet).UpdatedAt = fakeModel.UpdatedAt
			mdls[0].(*models.Faucet).CreatedAt = fakeModel.CreatedAt

			return mockQuery
		})
		mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(nil)

		faucet, err := s.dataAgent.FindOneByUUID(ctx, uuid, tenants)
		require.NoError(t, err)

		assert.Equal(t, expectedFaucet.CreatedAt, faucet.CreatedAt)
		assert.Equal(t, expectedFaucet.UpdatedAt, faucet.UpdatedAt)
		assert.Equal(t, expectedFaucet.Amount, faucet.Amount)
		assert.Equal(t, expectedFaucet.UUID, faucet.UUID)
		assert.Equal(t, expectedFaucet.MaxBalance, faucet.MaxBalance)
		assert.Equal(t, expectedFaucet.ChainRule, faucet.ChainRule)
		assert.Equal(t, expectedFaucet.TenantID, faucet.TenantID)
		assert.Equal(t, expectedFaucet.CreditorAccount, faucet.CreditorAccount)
		assert.Equal(t, expectedFaucet.Cooldown, faucet.Cooldown)
		assert.Equal(t, expectedFaucet.Name, faucet.Name)
	})

	s.T().Run("should fail with same error if Select fails with NotFoundError", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(expectedErr)

		account, err := s.dataAgent.FindOneByUUID(ctx, uuid, tenants)
		assert.Nil(t, account)
		assert.True(t, errors.IsNotFoundError(err))
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(expectedErr)

		account, err := s.dataAgent.FindOneByUUID(ctx, uuid, tenants)
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}
