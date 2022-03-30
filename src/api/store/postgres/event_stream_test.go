// +build unit

package postgres

import (
	"context"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
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
		assert.Equal(t, fakeEventStream.Specs, eventStream.Specs)
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
