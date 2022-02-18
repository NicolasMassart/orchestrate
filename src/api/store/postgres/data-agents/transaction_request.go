package dataagents

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store/models"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	gopg "github.com/go-pg/pg/v9"
)

const txRequestDAComponent = "data-agents.transaction-request"

// PGTransactionRequest is a transaction request data agent for PostgreSQL
type PGTransactionRequest struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGTransactionRequest creates a new PGTransactionRequest
func NewPGTransactionRequest(db pg.DB) store.TransactionRequestAgent {
	return &PGTransactionRequest{db: db, logger: log.NewLogger().SetComponent(txRequestDAComponent)}
}

// Insert Inserts a new transaction request in DB
func (agent *PGTransactionRequest) Insert(ctx context.Context, txRequest *entities.TxRequest, requestHash, scheduleUUID string) error {
	txRequest.CreatedAt = time.Now().UTC()
	txRequest.Params.CreatedAt = txRequest.CreatedAt
	txRequest.Params.UpdatedAt = txRequest.CreatedAt
	model := parsers.NewTxRequestModel(txRequest, requestHash)

	scheduleID, err := getScheduleIDByUUID(ctx, agent.db, agent.logger, scheduleUUID)
	if err != nil {
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}
	model.ScheduleID = &scheduleID

	err = pg.Insert(ctx, agent.db, model)
	if err != nil {
		errMessage := "error executing selectOrInsert"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.PostgresConnectionError(errMessage).ExtendComponent(txRequestDAComponent)
	}

	return nil
}

func (agent *PGTransactionRequest) FindOneByIdempotencyKey(ctx context.Context, idempotencyKey, tenantID, ownerID string) (*entities.TxRequest, error) {
	txRequest := &models.TransactionRequest{}
	query := agent.db.ModelContext(ctx, txRequest).
		Where("idempotency_key = ?", idempotencyKey).
		Relation("Schedule")

	query = query.Where("schedule.tenant_id = ?", tenantID)
	query = pg.WhereAllowedOwner(query, "schedule.owner_id", ownerID)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to fetch tx-request by idempotency key")
		}
		return nil, errors.FromError(err).ExtendComponent(txRequestDAComponent)
	}

	return parsers.NewTxRequestEntity(txRequest), nil
}

func (agent *PGTransactionRequest) FindOneByUUID(ctx context.Context, scheduleUUID string, tenants []string, ownerID string) (*entities.TxRequest, error) {
	txRequest := &models.TransactionRequest{}
	query := agent.db.ModelContext(ctx, txRequest).
		Where("schedule.uuid = ?", scheduleUUID).
		Relation("Schedule")

	query = pg.WhereAllowedTenants(query, "schedule.tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "schedule.owner_id", ownerID)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to find tx-request by uuid")
		}
		return nil, errors.FromError(err).ExtendComponent(txRequestDAComponent)
	}

	return parsers.NewTxRequestEntity(txRequest), nil
}

func (agent *PGTransactionRequest) Search(ctx context.Context, filters *entities.TransactionRequestFilters, tenants []string, ownerID string) ([]*entities.TxRequest, error) {
	var txRequests []*models.TransactionRequest

	query := agent.db.ModelContext(ctx, &txRequests).Relation("Schedule")

	if len(filters.IdempotencyKeys) > 0 {
		query = query.Where("transaction_request.idempotency_key in (?)", gopg.In(filters.IdempotencyKeys))
	}

	query = pg.WhereAllowedTenants(query, "schedule.tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "schedule.owner_id", ownerID)

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to fetch tx-requests")
		}
		return nil, errors.FromError(err).ExtendComponent(txRequestDAComponent)
	}

	return parsers.NewTxRequestEntityArr(txRequests), nil
}
