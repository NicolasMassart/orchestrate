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

type pgLogTestSuite struct {
	suite.Suite
	mockPGClient *mocks.MockClient
	dataAgent    *PGLog
}

func TestPGLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := new(pgLogTestSuite)
	s.mockPGClient = mocks.NewMockClient(ctrl)
	s.dataAgent = NewPGLog(s.mockPGClient)

	suite.Run(t, s)
}

func (s *pgLogTestSuite) TestInsert() {
	ctx := context.Background()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	uuid := "uuid"

	s.T().Run("should insert and parse successfully", func(t *testing.T) {
		fakeLog := testdata.FakeLog()
		mockQuery := mocks.NewMockQuery(ctrl)
		jobID := 666

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
				mdls[0].(*models.Job).ID = jobID
				return mockQuery
			}),
			mockQuery.EXPECT().Column("id").Return(mockQuery),
			mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery),
			mockQuery.EXPECT().Select().Return(nil),
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, mdls ...interface{}) postgres.Query {
				logModel := mdls[0].(*models.Log)
				assert.NotEmpty(t, logModel.UUID)
				assert.Equal(t, fakeLog.Message, logModel.Message)
				assert.Equal(t, fakeLog.Status.String(), logModel.Status)
				assert.Equal(t, jobID, *logModel.JobID)

				return mockQuery
			}),
		)
		mockQuery.EXPECT().Insert().Return(nil)

		log, err := s.dataAgent.Insert(ctx, fakeLog, uuid)
		require.NoError(t, err)

		assert.False(t, log.CreatedAt.IsZero())
		assert.Equal(t, fakeLog.Message, log.Message)
		assert.Equal(t, fakeLog.Status, log.Status)
	})

	s.T().Run("should fail with same error if Select fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Column("id").Return(mockQuery),
			mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery),
			mockQuery.EXPECT().Select().Return(expectedErr),
		)

		log, err := s.dataAgent.Insert(ctx, testdata.FakeLog(), uuid)
		require.Nil(t, log)
		require.True(t, errors.IsPostgresConnectionError(err))
	})

	s.T().Run("should fail with same error if Insert fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")
		mockQuery := mocks.NewMockQuery(ctrl)

		gomock.InOrder(
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Column("id").Return(mockQuery),
			mockQuery.EXPECT().Where("uuid = ?", uuid).Return(mockQuery),
			mockQuery.EXPECT().Select().Return(nil),
			s.mockPGClient.EXPECT().ModelContext(ctx, gomock.Any()).Return(mockQuery),
			mockQuery.EXPECT().Insert().Return(expectedErr),
		)

		log, err := s.dataAgent.Insert(ctx, testdata.FakeLog(), uuid)
		require.Nil(t, log)
		require.True(t, errors.IsPostgresConnectionError(err))
	})
}
