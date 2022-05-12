//go:build unit
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
)

type pgNotificationTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGNotification
}

func TestPGNotification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgNotificationTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGNotification(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgNotificationTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeNotif := testdata.FakeNotification()
		fakeNotif.UUID = ""
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			notifModel := mdls[0].(*models.Notification)
			assert.NotEmpty(t, notifModel.UUID)
			assert.Equal(t, fakeNotif.APIVersion, notifModel.APIVersion)
			assert.Equal(t, fakeNotif.Error, notifModel.Error)
			assert.Equal(t, fakeNotif.SourceUUID, notifModel.SourceUUID)
			assert.Equal(t, fakeNotif.SourceType.String(), notifModel.SourceType)
			assert.Equal(t, fakeNotif.Type.String(), notifModel.Type)
			assert.Equal(t, fakeNotif.Status.String(), notifModel.Status)

			return mockQuery
		})
		mockQuery.EXPECT().Insert().Return(nil)

		notif, err := s.dataAgent.Insert(ctx, fakeNotif)
		require.NoError(t, err)

		assert.False(t, notif.CreatedAt.IsZero())
		assert.False(t, notif.UpdatedAt.IsZero())
		assert.NotEmpty(t, notif.UUID)
		assert.Equal(t, fakeNotif.APIVersion, notif.APIVersion)
		assert.Equal(t, fakeNotif.Error, notif.Error)
		assert.Equal(t, fakeNotif.SourceUUID, notif.SourceUUID)
		assert.Equal(t, fakeNotif.SourceType, notif.SourceType)
		assert.Equal(t, fakeNotif.Type, notif.Type)
		assert.Equal(t, fakeNotif.Status, notif.Status)
	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Insert().Return(expectedErr)

		account, err := s.dataAgent.Insert(ctx, testdata.FakeNotification())
		assert.Nil(t, account)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}

func (s *pgNotificationTestSuite) TestUpdate() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.T().Run("should update and parse successfully", func(t *testing.T) {
		fakeNotif := testdata.FakeNotification()
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
			notifModel := mdls[0].(*models.Notification)
			assert.Equal(t, fakeNotif.UUID, notifModel.UUID)
			assert.Equal(t, fakeNotif.Status.String(), notifModel.Status)
			assert.Equal(t, fakeNotif.Error, notifModel.Error)
			assert.Equal(t, fakeNotif.SourceUUID, notifModel.SourceUUID)
			assert.Equal(t, fakeNotif.APIVersion, notifModel.APIVersion)
			assert.Equal(t, fakeNotif.Type.String(), notifModel.Type)
			assert.Equal(t, fakeNotif.SourceType.String(), notifModel.SourceType)

			return mockQuery
		})
		mockQuery.EXPECT().Where("uuid = ?", fakeNotif.UUID).Return(mockQuery)
		mockQuery.EXPECT().UpdateNotZero().Return(nil)

		notif, err := s.dataAgent.Update(ctx, fakeNotif)
		require.NoError(t, err)

		assert.False(t, notif.UpdatedAt.Equal(notif.CreatedAt))
		assert.Equal(t, fakeNotif.UUID, notif.UUID)
		assert.Equal(t, fakeNotif.Status, notif.Status)
		assert.Equal(t, fakeNotif.Error, notif.Error)
		assert.Equal(t, fakeNotif.SourceUUID, notif.SourceUUID)
		assert.Equal(t, fakeNotif.APIVersion, notif.APIVersion)
		assert.Equal(t, fakeNotif.Type, notif.Type)
		assert.Equal(t, fakeNotif.SourceType, notif.SourceType)
	})

	s.T().Run("should fail with same error if UpdateNotZero fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().Where("uuid = ?", gomock.Any()).Return(mockQuery)
		mockQuery.EXPECT().UpdateNotZero().Return(expectedErr)

		notif, err := s.dataAgent.Update(ctx, testdata.FakeNotification())
		assert.Nil(t, notif)
		assert.True(t, errors.IsPostgresConnectionError(err))
	})
}
