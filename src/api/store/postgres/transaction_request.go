package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/go-pg/pg/v10"

	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store/models"
)

type PGTransactionRequest struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.TransactionRequestAgent = &PGTransactionRequest{}

func NewPGTransactionRequest(client postgres.Client) *PGTransactionRequest {
	return &PGTransactionRequest{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.transaction-request"),
	}
}

func (agent *PGTransactionRequest) Insert(ctx context.Context, txRequest *entities.TxRequest, requestHash, scheduleUUID string) (*entities.TxRequest, error) {
	txRequest.CreatedAt = time.Now().UTC()
	txRequest.Params.CreatedAt = txRequest.CreatedAt
	txRequest.Params.UpdatedAt = txRequest.CreatedAt

	model := models.NewTxRequest(txRequest, requestHash)

	scheduleID, err := getScheduleIDByUUID(ctx, agent.client, agent.logger, scheduleUUID)
	if err != nil {
		return nil, err
	}
	model.ScheduleID = &scheduleID

	err = agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMessage := "failed to insert transaction request"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.PostgresConnectionError(errMessage).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGTransactionRequest) FindOneByIdempotencyKey(ctx context.Context, idempotencyKey, tenantID, ownerID string) (*entities.TxRequest, error) {
	txRequest := &models.TransactionRequest{}

	err := agent.client.ModelContext(ctx, txRequest).
		Where("idempotency_key = ?", idempotencyKey).
		Relation("Schedule").
		Where("schedule.tenant_id = ?", tenantID).
		WhereAllowedOwner("schedule.owner_id", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("transaction request not found by idempotency key")
		}

		errMessage := "failed to fetch tx-request by idempotency key"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return txRequest.ToEntity(), nil
}

func (agent *PGTransactionRequest) FindOneByUUID(ctx context.Context, scheduleUUID string, tenants []string, ownerID string) (*entities.TxRequest, error) {
	txRequest := &models.TransactionRequest{}

	err := agent.client.ModelContext(ctx, txRequest).
		Where("schedule.uuid = ?", scheduleUUID).
		Relation("Schedule").
		WhereAllowedTenants("schedule.tenant_id", tenants).
		WhereAllowedOwner("schedule.owner_id", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("transaction request not found by uuid")
		}

		errMessage := "failed to find tx-request by uuid"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return txRequest.ToEntity(), nil
}

func (agent *PGTransactionRequest) Search(ctx context.Context, filters *entities.TransactionRequestFilters, tenants []string, ownerID string) ([]*entities.TxRequest, error) {
	var txRequests []*models.TransactionRequest

	q := agent.client.ModelContext(ctx, &txRequests).Relation("Schedule")

	if len(filters.IdempotencyKeys) > 0 {
		q = q.Where("transaction_request.idempotency_key in (?)", pg.In(filters.IdempotencyKeys))
	}

	err := q.WhereAllowedTenants("schedule.tenant_id", tenants).
		WhereAllowedOwner("schedule.owner_id", ownerID).
		Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMessage := "failed to fetch tx-requests"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return models.NewTxRequests(txRequests), nil
}
