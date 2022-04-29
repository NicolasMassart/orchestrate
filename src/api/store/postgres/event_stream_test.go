//go:build unit
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

type pgEventStreamTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGEventStream
}

func TestPGEventStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgEventStreamTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGEventStream(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgEventStreamTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeEventStream := testdata.FakeWebhookEventStream()
		fakeEventStream.UUID = ""
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(nil)

		eventStream, err := s.dataAgent.Insert(ctx, fakeEventStream)
		require.NoError(t, err)

		assert.False(t, eventStream.CreatedAt.IsZero())
		assert.False(t, eventStream.UpdatedAt.IsZero())
		assert.NotEmpty(t, eventStream.UUID)
		assert.Equal(t, fakeEventStream.Name, eventStream.Name)
		assert.Equal(t, fakeEventStream.ChainUUID, eventStream.ChainUUID)
		assert.Equal(t, fakeEventStream.Status, eventStream.Status)
		assert.Equal(t, fakeEventStream.Channel, eventStream.Channel)
		assert.Equal(t, fakeEventStream.TenantID, eventStream.TenantID)
		assert.Equal(t, fakeEventStream.OwnerID, eventStream.OwnerID)
		assert.Equal(t, fakeEventStream.Labels, eventStream.Labels)
		assert.Equal(t, fakeEventStream.Webhook, eventStream.Webhook)
	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(expectedErr)

		eventStream, err := s.dataAgent.Insert(ctx, testdata.FakeWebhookEventStream())
		assert.Nil(t, eventStream)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgEventStreamTestSuite) TestSearch() {
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

		_, err := s.dataAgent.Search(ctx, &entities.EventStreamFilters{}, tenants, owner)
		require.NoError(t, err)
	})

	s.T().Run("should search successfully with filters", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		filters := &entities.EventStreamFilters{
			Names:    []string{"name1", "name2"},
			TenantID: "tenant",
		}

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		gomock.InOrder(
			mockQuery.EXPECT().Where("name in (?)", gomock.Any()).Return(mockQuery),
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

		accounts, err := s.dataAgent.Search(ctx, &entities.EventStreamFilters{}, tenants, owner)
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

		accounts, err := s.dataAgent.Search(ctx, &entities.EventStreamFilters{}, tenants, owner)
		assert.Empty(t, accounts)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgEventStreamTestSuite) TestFindOneByUUID() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	tenants := []string{"tenant"}
	owner := "owner"
	uuid := "uuid"

	s.T().Run("should find one successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)
		expectedEventStream := testdata.FakeWebhookEventStream()

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			assert.Equal(t, &models.EventStream{}, mdls[0].(*models.EventStream))

			fakeModel := models.NewEventStream(expectedEventStream)
			mdls[0].(*models.EventStream).ChainUUID = fakeModel.ChainUUID
			mdls[0].(*models.EventStream).UUID = fakeModel.UUID
			mdls[0].(*models.EventStream).Channel = fakeModel.Channel
			mdls[0].(*models.EventStream).Specs = fakeModel.Specs
			mdls[0].(*models.EventStream).Status = fakeModel.Status
			mdls[0].(*models.EventStream).Labels = fakeModel.Labels
			mdls[0].(*models.EventStream).Name = fakeModel.Name
			mdls[0].(*models.EventStream).TenantID = fakeModel.TenantID
			mdls[0].(*models.EventStream).UpdatedAt = fakeModel.UpdatedAt
			mdls[0].(*models.EventStream).CreatedAt = fakeModel.CreatedAt

			return mockQuery
		})
		mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(nil)

		es, err := s.dataAgent.FindOneByUUID(ctx, uuid, tenants, owner)
		require.NoError(t, err)

		assert.Equal(t, expectedEventStream.CreatedAt, es.CreatedAt)
		assert.Equal(t, expectedEventStream.UpdatedAt, es.UpdatedAt)
		assert.Equal(t, expectedEventStream.Webhook, es.Webhook)
		assert.Equal(t, expectedEventStream.UUID, es.UUID)
		assert.Equal(t, expectedEventStream.Status, es.Status)
		assert.Equal(t, expectedEventStream.TenantID, es.TenantID)
		assert.Equal(t, expectedEventStream.Labels, es.Labels)
		assert.Equal(t, expectedEventStream.Channel, es.Channel)
		assert.Equal(t, expectedEventStream.Name, es.Name)
	})

	s.T().Run("should fail with same error if Select fails with NotFoundError", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(expectedErr)

		account, err := s.dataAgent.FindOneByUUID(ctx, uuid, tenants, owner)
		assert.Nil(t, account)
		assert.True(t, errors.IsNotFoundError(err))
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().SelectOne().Return(expectedErr)

		account, err := s.dataAgent.FindOneByUUID(ctx, uuid, tenants, owner)
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgEventStreamTestSuite) TestDelete() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	tenants := []string{"tenant"}
	owner := "owner"

	s.T().Run("should delete successfully", func(t *testing.T) {
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", "uuid").Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().Delete().Return(nil)

		err := s.dataAgent.Delete(ctx, "uuid", tenants, owner)
		require.NoError(t, err)
	})

	s.T().Run("should fail with same error if Delete fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", "uuid").Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedTenants("", tenants).Return(mockQuery)
		mockQuery.EXPECT().WhereAllowedOwner("", owner).Return(mockQuery)
		mockQuery.EXPECT().Delete().Return(expectedErr)

		err := s.dataAgent.Delete(ctx, "uuid", tenants, owner)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}
