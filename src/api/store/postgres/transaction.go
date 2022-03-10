package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/api/store"

	"github.com/consensys/orchestrate/src/infra/postgres"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"
)

type PGTransaction struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.TransactionAgent = &PGTransaction{}

func NewPGTransaction(client postgres.Client) *PGTransaction {
	return &PGTransaction{
		client: client,
		logger: log.NewLogger().SetComponent("data-agent.transaction"),
	}
}

func (agent *PGTransaction) Insert(ctx context.Context, tx *entities.ETHTransaction) (*entities.ETHTransaction, error) {
	model := models.NewTransaction(tx)
	model.UUID = uuid.Must(uuid.NewV4()).String()
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMessage := "failed to insert transaction"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGTransaction) Update(ctx context.Context, tx *entities.ETHTransaction, jobUUID string) (*entities.ETHTransaction, error) {
	txModel := &models.Transaction{}
	err := agent.client.ModelContext(ctx, txModel).
		Join("JOIN jobs AS j ON j.transaction_id = transaction.id").
		Where("j.uuid = ?", jobUUID).
		SelectOne()
	if err != nil {
		errMessage := "failed to find transaction by job uuid"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	model := models.NewTransaction(tx)
	model.UpdatedAt = time.Now().UTC()

	err = agent.client.ModelContext(ctx, model).Where("uuid = ?", txModel.UUID).Update()
	if err != nil {
		errMsg := "failed to update transaction"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ToEntity(), nil
}
