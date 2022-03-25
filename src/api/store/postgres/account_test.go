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

type pgAccountTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGAccount
}

func TestPGAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgAccountTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGAccount(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgAccountTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeAccount := testdata.FakeAccount()
		fakeAccount.CreatedAt = time.Time{}
		fakeAccount.UpdatedAt = time.Time{}
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			accModel := mdls[0].(*models.Account)
			assert.Equal(t, fakeAccount.Alias, accModel.Alias)
			assert.Equal(t, fakeAccount.Address.Hex(), accModel.Address)
			assert.Equal(t, fakeAccount.PublicKey.String(), accModel.PublicKey)
			assert.Equal(t, fakeAccount.CompressedPublicKey.String(), accModel.CompressedPublicKey)
			assert.Equal(t, fakeAccount.TenantID, accModel.TenantID)
			assert.Equal(t, fakeAccount.OwnerID, accModel.OwnerID)
			assert.Equal(t, fakeAccount.StoreID, accModel.StoreID)
			assert.Equal(t, fakeAccount.Attributes, accModel.Attributes)

			return mockQuery
		})
		mockQuery.EXPECT().Insert().Return(nil)

		account, err := s.dataAgent.Insert(ctx, fakeAccount)
		require.NoError(t, err)

		assert.False(t, account.CreatedAt.IsZero())
		assert.False(t, account.UpdatedAt.IsZero())
		assert.Equal(t, fakeAccount.Alias, account.Alias)
		assert.Equal(t, fakeAccount.Address, account.Address)
		assert.Equal(t, fakeAccount.PublicKey, account.PublicKey)
		assert.Equal(t, fakeAccount.CompressedPublicKey, account.CompressedPublicKey)
		assert.Equal(t, fakeAccount.TenantID, account.TenantID)
		assert.Equal(t, fakeAccount.OwnerID, account.OwnerID)
		assert.Equal(t, fakeAccount.StoreID, account.StoreID)
		assert.Equal(t, fakeAccount.Attributes, account.Attributes)
	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(expectedErr)

		account, err := s.dataAgent.Insert(ctx, testdata.FakeAccount())
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgAccountTestSuite) TestUpdate() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should update and parse successfully", func(t *testing.T) {
		fakeAccount := testdata.FakeAccount()
		fakeAccount.UpdatedAt = time.Time{}
		fakeAccount.CreatedAt = time.Time{}
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			accModel := mdls[0].(*models.Account)
			assert.Equal(t, fakeAccount.Alias, accModel.Alias)
			assert.Equal(t, fakeAccount.Address.Hex(), accModel.Address)
			assert.Equal(t, fakeAccount.PublicKey.String(), accModel.PublicKey)
			assert.Equal(t, fakeAccount.CompressedPublicKey.String(), accModel.CompressedPublicKey)
			assert.Equal(t, fakeAccount.TenantID, accModel.TenantID)
			assert.Equal(t, fakeAccount.OwnerID, accModel.OwnerID)
			assert.Equal(t, fakeAccount.StoreID, accModel.StoreID)
			assert.Equal(t, fakeAccount.Attributes, accModel.Attributes)

			return mockQuery
		})
		gomock.InOrder(
			mockQuery.EXPECT().Where("address = ?", fakeAccount.Address.Hex()).Return(mockQuery),
			mockQuery.EXPECT().Where("tenant_id = ?", fakeAccount.TenantID).Return(mockQuery),
			mockQuery.EXPECT().Where("owner_id = ?", fakeAccount.OwnerID).Return(mockQuery),
		)
		mockQuery.EXPECT().UpdateNotZero().Return(nil)

		account, err := s.dataAgent.Update(ctx, fakeAccount)
		require.NoError(t, err)

		assert.True(t, account.CreatedAt.IsZero())
		assert.False(t, account.UpdatedAt.IsZero())
		assert.Equal(t, fakeAccount.Alias, account.Alias)
		assert.Equal(t, fakeAccount.Address, account.Address)
		assert.Equal(t, fakeAccount.PublicKey, account.PublicKey)
		assert.Equal(t, fakeAccount.CompressedPublicKey, account.CompressedPublicKey)
		assert.Equal(t, fakeAccount.TenantID, account.TenantID)
		assert.Equal(t, fakeAccount.OwnerID, account.OwnerID)
		assert.Equal(t, fakeAccount.StoreID, account.StoreID)
		assert.Equal(t, fakeAccount.Attributes, account.Attributes)
	})

	s.T().Run("should update without owner if not specified", func(t *testing.T) {
		fakeAccount := testdata.FakeAccount()
		fakeAccount.OwnerID = ""
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		gomock.InOrder(
			mockQuery.EXPECT().Where("address = ?", fakeAccount.Address.Hex()).Return(mockQuery),
			mockQuery.EXPECT().Where("tenant_id = ?", fakeAccount.TenantID).Return(mockQuery),
		)
		mockQuery.EXPECT().UpdateNotZero().Return(nil)

		account, err := s.dataAgent.Update(ctx, fakeAccount)
		require.NoError(t, err)
		require.NotNil(t, account)
	})

	s.T().Run("should fail with same error if UpdateNotZero fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where(gomock.Any(), gomock.Any()).Return(mockQuery).Times(3)
		mockQuery.EXPECT().UpdateNotZero().Return(expectedErr)

		account, err := s.dataAgent.Update(ctx, testdata.FakeAccount())
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgAccountTestSuite) TestSearch() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}
	owner := "owner"

	s.T().Run("should search successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().Order("id ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)

		_, err := s.dataAgent.Search(ctx, &entities.AccountFilters{}, tenants, owner)
		require.NoError(t, err)
	})

	s.T().Run("should search successfully with filters", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		filters := &entities.AccountFilters{
			Aliases:  []string{"alias1", "alias2"},
			TenantID: "tenant",
		}

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		gomock.InOrder(
			mockQuery.EXPECT().Where("alias in (?)", gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Where("tenant_id = ?", filters.TenantID).Return(mockQuery),
		)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().Order("id ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(nil)

		_, err := s.dataAgent.Search(ctx, filters, tenants, owner)
		require.NoError(t, err)
	})

	s.T().Run("should return empty array if Select fails with NotFoundError", func(t *testing.T) {
		notFoundErr := errors.NotFoundError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().Order("id ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(notFoundErr)

		accounts, err := s.dataAgent.Search(ctx, &entities.AccountFilters{}, tenants, owner)
		require.NoError(t, err)

		assert.Empty(t, accounts)
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().Order("id ASC").Return(mockQuery)
		mockQuery.EXPECT().Select().Return(expectedErr)

		accounts, err := s.dataAgent.Search(ctx, &entities.AccountFilters{}, tenants, owner)
		assert.Empty(t, accounts)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgAccountTestSuite) TestFindOneByAddress() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}
	owner := "owner"

	s.T().Run("should find one successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		address := "0xB104c3bcdA6eeE05b4A3cE25b97B7501EbCbb0A9"
		expectedAccount := testdata.FakeAccount()

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			assert.Equal(t, &models.Account{}, mdls[0].(*models.Account))

			fakeModel := models.NewAccount(expectedAccount)
			mdls[0].(*models.Account).PublicKey = fakeModel.PublicKey
			mdls[0].(*models.Account).CompressedPublicKey = fakeModel.CompressedPublicKey
			mdls[0].(*models.Account).Address = fakeModel.Address
			mdls[0].(*models.Account).Alias = fakeModel.Alias
			mdls[0].(*models.Account).Attributes = fakeModel.Attributes
			mdls[0].(*models.Account).OwnerID = fakeModel.OwnerID
			mdls[0].(*models.Account).StoreID = fakeModel.StoreID
			mdls[0].(*models.Account).TenantID = fakeModel.TenantID
			mdls[0].(*models.Account).UpdatedAt = fakeModel.UpdatedAt
			mdls[0].(*models.Account).CreatedAt = fakeModel.CreatedAt

			return mockQuery
		})
		mockQuery.EXPECT().Where("address = ?", address).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(nil)

		account, err := s.dataAgent.FindOneByAddress(ctx, address, tenants, owner)
		require.NoError(t, err)

		assert.Equal(t, expectedAccount.CreatedAt, account.CreatedAt)
		assert.Equal(t, expectedAccount.UpdatedAt, account.UpdatedAt)
		assert.Equal(t, expectedAccount.Alias, account.Alias)
		assert.Equal(t, expectedAccount.Address, account.Address)
		assert.Equal(t, expectedAccount.PublicKey, account.PublicKey)
		assert.Equal(t, expectedAccount.CompressedPublicKey, account.CompressedPublicKey)
		assert.Equal(t, expectedAccount.TenantID, account.TenantID)
		assert.Equal(t, expectedAccount.OwnerID, account.OwnerID)
		assert.Equal(t, expectedAccount.StoreID, account.StoreID)
		assert.Equal(t, expectedAccount.Attributes, account.Attributes)
	})

	s.T().Run("should fail with same error if Select fails with NotFoundError", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		mockQuery := mocks.NewMockQuery(ctrl)
		address := "0xB104c3bcdA6eeE05b4A3cE25b97B7501EbCbb0A9"

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("address = ?", address).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(expectedErr)

		account, err := s.dataAgent.FindOneByAddress(ctx, address, tenants, owner)
		assert.Nil(t, account)
		assert.True(t, errors.IsNotFoundError(err))
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)
		address := "0xB104c3bcdA6eeE05b4A3cE25b97B7501EbCbb0A9"

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("address = ?", address).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(expectedErr)

		account, err := s.dataAgent.FindOneByAddress(ctx, address, tenants, owner)
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}
