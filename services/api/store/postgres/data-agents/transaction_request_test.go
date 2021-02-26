// +build unit
// +build !race
// +build !integration

package dataagents

import (
	"context"
	"github.com/ConsenSys/orchestrate/pkg/multitenancy"
	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	"testing"

	"github.com/gofrs/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	pgTestUtils "github.com/ConsenSys/orchestrate/pkg/database/postgres/testutils"
	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/store/models"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
	"github.com/ConsenSys/orchestrate/services/api/store/postgres/migrations"
)

type txRequestTestSuite struct {
	suite.Suite
	agents *PGAgents
	pg     *pgTestUtils.PGTestHelper
}

func TestPGTransactionRequest(t *testing.T) {
	s := new(txRequestTestSuite)
	suite.Run(t, s)
}

func (s *txRequestTestSuite) SetupSuite() {
	s.pg, _ = pgTestUtils.NewPGTestHelper(nil, migrations.Collection)
	s.pg.InitTestDB(s.T())
}

func (s *txRequestTestSuite) SetupTest() {
	s.pg.UpgradeTestDB(s.T())
	s.agents = New(s.pg.DB)
}

func (s *txRequestTestSuite) TearDownTest() {
	s.pg.DowngradeTestDB(s.T())
}

func (s *txRequestTestSuite) TearDownSuite() {
	s.pg.DropTestDB(s.T())
}

func (s *txRequestTestSuite) TestPGTransactionRequest_Insert() {
	ctx := context.Background()

	s.T().Run("should insert model successfully if uuid is not defined", func(t *testing.T) {
		txRequest := testutils.FakeTxRequest(0)
		err := insertTxRequest(ctx, s.agents, txRequest)

		assert.NoError(t, err)
		assert.NotEmpty(t, txRequest.ID)
		assert.NotEmpty(t, txRequest.Schedule.UUID)
	})

	s.T().Run("should insert model successfully if uuid is already set", func(t *testing.T) {
		txRequest := testutils.FakeTxRequest(0)
		txRequestUUID := txRequest.Schedule.UUID

		err := insertTxRequest(ctx, s.agents, txRequest)

		assert.NoError(t, err)
		assert.NotEmpty(t, txRequest.ID)
		assert.Equal(t, txRequestUUID, txRequest.Schedule.UUID)
	})
	
	s.T().Run("should insert model successfully if idempotencyKey is empty", func(t *testing.T) {
		txRequest := testutils.FakeTxRequest(0)

		err := insertTxRequest(ctx, s.agents, txRequest)

		assert.NoError(t, err)
		assert.NotEmpty(t, txRequest.ID)
	})
}

func (s *txRequestTestSuite) TestPGTransactionRequest_FindOneByIdempotencyKey() {
	ctx := context.Background()
	txRequest := testutils.FakeTxRequest(0)
	err := insertTxRequest(ctx, s.agents, txRequest)
	assert.NoError(s.T(), err)

	s.T().Run("should find request successfully", func(t *testing.T) {
		txRequestRetrieved, err := s.agents.TransactionRequest().
			FindOneByIdempotencyKey(ctx, txRequest.IdempotencyKey, txRequest.Schedule.TenantID)

		assert.NoError(t, err)
		assert.Equal(t, txRequest.IdempotencyKey, txRequestRetrieved.IdempotencyKey)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestRetrieved.Schedule.UUID)
	})

	s.T().Run("should return NotFoundError if request is not found", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().FindOneByIdempotencyKey(ctx, txRequest.IdempotencyKey, "randomTenant")
		assert.True(t, errors.IsNotFoundError(err))
	})

	s.T().Run("should return NotFoundError if request is not found", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().FindOneByIdempotencyKey(ctx, "notExisting", txRequest.Schedule.TenantID)
		assert.True(t, errors.IsNotFoundError(err))
	})
}

func (s *txRequestTestSuite) TestPGTransactionRequest_FindOneByUUID() {
	ctx := context.Background()
	txRequest := testutils.FakeTxRequest(0)
	err := insertTxRequest(ctx, s.agents, txRequest)
	assert.Nil(s.T(), err)

	s.T().Run("should find request successfully for empty tenant", func(t *testing.T) {
		txRequestRetrieved, err := s.agents.TransactionRequest().FindOneByUUID(ctx, txRequest.Schedule.UUID, []string{multitenancy.Wildcard})

		assert.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestRetrieved.Schedule.UUID)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestRetrieved.Schedule.UUID)
	})

	s.T().Run("should find request successfully for default tenant", func(t *testing.T) {
		txRequestRetrieved, err := s.agents.TransactionRequest().FindOneByUUID(ctx, txRequest.Schedule.UUID, []string{multitenancy.DefaultTenant})

		assert.NoError(t, err)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestRetrieved.Schedule.UUID)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestRetrieved.Schedule.UUID)
	})

	s.T().Run("should return NotFoundError if uuid is not found", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().FindOneByUUID(ctx, uuid.Must(uuid.NewV4()).String(), []string{txRequest.Schedule.TenantID})
		assert.True(t, errors.IsNotFoundError(err))
	})

	s.T().Run("should return NotFoundError if tenant is not found", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().FindOneByUUID(ctx, txRequest.Schedule.UUID, []string{"notExisting"})
		assert.True(t, errors.IsNotFoundError(err))
	})
}

func (s *txRequestTestSuite) TestPGTransactionRequest_Search() {
	ctx := context.Background()
	txRequest := testutils.FakeTxRequest(0)
	err := insertTxRequest(ctx, s.agents, txRequest)
	assert.Nil(s.T(), err)

	s.T().Run("should find requests successfully", func(t *testing.T) {
		filter := &entities.TransactionRequestFilters{}
		txRequestsRetrieved, err := s.agents.TransactionRequest().Search(ctx, filter, []string{multitenancy.Wildcard})

		assert.NoError(t, err)
		assert.Len(t, txRequestsRetrieved, 1)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestsRetrieved[0].Schedule.UUID)
	})

	s.T().Run("should find requests successfully by idempotency keys", func(t *testing.T) {
		filter := &entities.TransactionRequestFilters{
			IdempotencyKeys: []string{txRequest.IdempotencyKey},
		}
		txRequestsRetrieved, err := s.agents.TransactionRequest().Search(ctx, filter, []string{multitenancy.Wildcard})

		assert.NoError(t, err)
		assert.Len(t, txRequestsRetrieved, 1)
		assert.Equal(t, txRequest.Schedule.UUID, txRequestsRetrieved[0].Schedule.UUID)
	})

	s.T().Run("should return empty array if nothing found in filter", func(t *testing.T) {
		filter := &entities.TransactionRequestFilters{
			IdempotencyKeys: []string{"notExisting"},
		}

		result, err := s.agents.TransactionRequest().Search(ctx, filter, []string{multitenancy.Wildcard})

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	s.T().Run("should return empty array if tenant is not found", func(t *testing.T) {
		filter := &entities.TransactionRequestFilters{}
		result, err := s.agents.TransactionRequest().Search(ctx, filter, []string{"NotExistingTenant"})

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func (s *txRequestTestSuite) TestPGTransactionRequest_ConnectionErr() {
	ctx := context.Background()

	// We drop the DB to make the test fail
	s.pg.DropTestDB(s.T())
	txRequest := testutils.FakeTxRequest(0)

	s.T().Run("should return PostgresConnectionError if insert fails", func(t *testing.T) {
		err := insertTxRequest(ctx, s.agents, txRequest)
		assert.True(t, errors.IsInternalError(err))
	})

	s.T().Run("should return PostgresConnectionError if find fails", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().FindOneByIdempotencyKey(ctx, txRequest.IdempotencyKey, txRequest.Schedule.TenantID)
		assert.True(t, errors.IsInternalError(err))
	})

	s.T().Run("should return PostgresConnectionError if find fails", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().FindOneByUUID(ctx, txRequest.Schedule.UUID, []string{multitenancy.Wildcard})
		assert.True(t, errors.IsInternalError(err))
	})

	s.T().Run("should return PostgresConnectionError if find fails", func(t *testing.T) {
		_, err := s.agents.TransactionRequest().Search(ctx, &entities.TransactionRequestFilters{}, []string{"tenant"})
		assert.True(t, errors.IsInternalError(err))
	})

	// We bring it back up
	s.pg.InitTestDB(s.T())
}

func insertTxRequest(ctx context.Context, agents *PGAgents, txReq *models.TransactionRequest) error {
	if err := agents.Schedule().Insert(ctx, txReq.Schedule); err != nil {
		return err
	}

	txReq.ScheduleID = &txReq.Schedule.ID
	if err := agents.TransactionRequest().Insert(ctx, txReq); err != nil {
		return err
	}

	return nil
}
