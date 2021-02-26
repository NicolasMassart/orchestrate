// +build unit
// +build !race
// +build !integration

package dataagents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	pgTestUtils "github.com/ConsenSys/orchestrate/pkg/database/postgres/testutils"
	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/store/models"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
	"github.com/ConsenSys/orchestrate/services/api/store/postgres/migrations"
)

type scheduleTestSuite struct {
	suite.Suite
	agents *PGAgents
	pg     *pgTestUtils.PGTestHelper
}

func TestPGSchedule(t *testing.T) {
	s := new(scheduleTestSuite)
	suite.Run(t, s)
}

func (s *scheduleTestSuite) SetupSuite() {
	s.pg, _ = pgTestUtils.NewPGTestHelper(nil, migrations.Collection)
	s.pg.InitTestDB(s.T())
}

func (s *scheduleTestSuite) SetupTest() {
	s.pg.UpgradeTestDB(s.T())
	s.agents = New(s.pg.DB)
}

func (s *scheduleTestSuite) TearDownTest() {
	s.pg.DowngradeTestDB(s.T())
}

func (s *scheduleTestSuite) TearDownSuite() {
	s.pg.DropTestDB(s.T())
}

func (s *scheduleTestSuite) TestPGSchedule_Insert() {
	s.T().Run("should insert model successfully", func(t *testing.T) {
		schedule := testutils.FakeSchedule("")
		err := s.agents.Schedule().Insert(context.Background(), schedule)

		assert.NoError(t, err)
		assert.NotEmpty(t, schedule.ID)
	})

	s.T().Run("should insert model without UUID successfully", func(t *testing.T) {
		schedule := testutils.FakeSchedule("")
		err := s.agents.Schedule().Insert(context.Background(), schedule)

		assert.NoError(t, err)
		assert.NotEmpty(t, schedule.UUID)
		assert.NotEmpty(t, schedule.ID)
	})
}

func (s *scheduleTestSuite) TestPGSchedule_FindOneByUUID() {
	ctx := context.Background()
	tenantID := "tenantID"
	schedule := testutils.FakeSchedule(tenantID)
	err := insertSchedule(ctx, s.agents, schedule)
	assert.NoError(s.T(), err)

	s.T().Run("should get model successfully as tenant", func(t *testing.T) {
		scheduleRetrieved, err := s.agents.Schedule().FindOneByUUID(ctx, schedule.UUID, []string{tenantID})

		assert.NoError(t, err)
		assertEqualSchedule(t, schedule, scheduleRetrieved)
	})

	s.T().Run("should return NotFoundError if select fails", func(t *testing.T) {
		_, err := s.agents.Schedule().FindOneByUUID(ctx, "b6fe7a2a-1a4d-49ca-99d8-8a34aa495ef0", []string{tenantID})
		assert.True(t, errors.IsNotFoundError(err))
	})

	s.T().Run("should return NotFoundError if select fails", func(t *testing.T) {
		_, err := s.agents.Schedule().FindOneByUUID(ctx, "b6fe7a2a-1a4d-49ca-99d8-8a34aa495ef0", []string{"randomID"})
		assert.True(t, errors.IsNotFoundError(err))
	})
}

func (s *scheduleTestSuite) TestPGSchedule_FindAll() {
	ctx := context.Background()
	tenantID := "tenantID"
	tenantID2 := "tenantID2"
	schedules := []*models.Schedule{
		testutils.FakeSchedule(tenantID),
		testutils.FakeSchedule(tenantID),
		testutils.FakeSchedule(tenantID2),
	}

	var err error
	for _, schedule := range schedules {
		err = insertSchedule(ctx, s.agents, schedule)
		assert.NoError(s.T(), err)
	}

	s.T().Run("should get models successfully as tenant", func(t *testing.T) {
		schedulesRetrieved, err := s.agents.Schedule().FindAll(ctx, []string{tenantID})

		assert.NoError(t, err)
		assert.Equal(t, 2, len(schedulesRetrieved))
		for idx, scheduleRetrieved := range schedulesRetrieved {
			assertEqualSchedule(t, schedules[idx], scheduleRetrieved)
		}
	})

	s.T().Run("should return empty array if nothing is found", func(t *testing.T) {
		schedules, err := s.agents.Schedule().FindAll(ctx, []string{"randomID"})
		assert.NoError(t, err)
		assert.Empty(t, schedules)
	})
}

func (s *scheduleTestSuite) TestPGSchedule_ConnectionErr() {
	ctx := context.Background()

	// We drop the DB to make the test fail
	s.pg.DropTestDB(s.T())
	schedule := testutils.FakeSchedule("")

	s.T().Run("should return PostgresConnectionError if Insert fails", func(t *testing.T) {
		err := s.agents.Schedule().Insert(ctx, schedule)
		assert.True(t, errors.IsInternalError(err))
	})

	s.T().Run("should return PostgresConnectionError if FindOneByUUID fails", func(t *testing.T) {
		_, err := s.agents.Schedule().FindOneByUUID(ctx, schedule.UUID, []string{"_"})
		assert.True(t, errors.IsInternalError(err))
	})

	s.T().Run("should return PostgresConnectionError if FindAll fails", func(t *testing.T) {
		_, err := s.agents.Schedule().FindAll(ctx, []string{"_"})
		assert.True(t, errors.IsInternalError(err))
	})

	// We bring it back up
	s.pg.InitTestDB(s.T())
}

func assertEqualSchedule(t *testing.T, expected, actual *models.Schedule) {
	assert.NotEmpty(t, actual.ID)
	assert.Equal(t, expected.UUID, actual.UUID)
	assert.Equal(t, expected.CreatedAt, actual.CreatedAt)
	assert.Equal(t, len(expected.Jobs), len(actual.Jobs))
	if len(expected.Jobs) == len(actual.Jobs) {
		for idx := range expected.Jobs {
			assert.NotEmpty(t, actual.Jobs[idx].ID)
			assert.Equal(t, expected.Jobs[idx].UUID, actual.Jobs[idx].UUID)
			assert.Equal(t, expected.Jobs[idx].Type, actual.Jobs[idx].Type)
		}
	}
}

func insertSchedule(ctx context.Context, agents *PGAgents, schedule *models.Schedule) error {
	err := agents.Schedule().Insert(ctx, schedule)
	if err != nil {
		return err
	}

	for _, job := range schedule.Jobs {
		if _, err := agents.Chain().FindOneByUUID(ctx, job.ChainUUID, []string{}); errors.IsNotFoundError(err) {
			chain := testutils.FakeChainModel()
			chain.UUID = job.ChainUUID
			if err := agents.Chain().Insert(ctx, chain); err != nil {
				return err
			}
		}

		job.Schedule = schedule
		if err := agents.Transaction().Insert(ctx, job.Transaction); err != nil {
			return err
		}
		if err := agents.Job().Insert(ctx, job); err != nil {
			return err
		}
	}

	return nil
}
