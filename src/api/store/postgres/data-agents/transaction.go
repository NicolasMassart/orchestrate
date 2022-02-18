package dataagents

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/gofrs/uuid"
)

const txDAComponent = "data-agent.transaction"

// PGLog is a log data agent for PostgreSQL
type PGTransaction struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGLog creates a new PGLog
func NewPGTransaction(db pg.DB) store.TransactionAgent {
	return &PGTransaction{db: db, logger: log.NewLogger().SetComponent(txDAComponent)}
}

// Insert Inserts a new log in DB
func (agent *PGTransaction) Insert(ctx context.Context, tx *entities.ETHTransaction) error {
	model := parsers.NewTransactionModel(tx)
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	err := pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert transaction")
		return errors.FromError(err).ExtendComponent(txDAComponent)
	}

	utils.CopyPtr(parsers.NewTransactionEntity(model), tx)
	return nil
}

func (agent *PGTransaction) Update(ctx context.Context, tx *entities.ETHTransaction, jobUUID string) error {
	curTx, err := agent.FindOneByJobUUID(ctx, jobUUID)
	if err != nil || curTx == nil {
		errMsg := "failed to find transaction by jobUUID"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).ExtendComponent(txDAComponent)
	}

	model := parsers.NewTransactionModel(tx)
	model.UpdatedAt = time.Now().UTC()

	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", curTx.UUID)
	err = pg.Update(ctx, query)
	if err != nil {
		errMsg := "failed to update transaction"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).ExtendComponent(txDAComponent)
	}

	utils.CopyPtr(parsers.NewTransactionEntity(model), tx)
	return nil
}

func (agent *PGTransaction) FindOneByJobUUID(ctx context.Context, jobUUID string) (*entities.ETHTransaction, error) {
	model := &models.Transaction{}
	query := agent.db.ModelContext(ctx, model).
		Join("JOIN jobs AS j ON j.transaction_id = transaction.id").
		Where("j.uuid = ?", jobUUID)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to find transaction by job uuid")
		return nil, errors.FromError(err).ExtendComponent(txDAComponent)
	}

	return parsers.NewTransactionEntity(model), nil
}

func getTxIDByUUID(ctx context.Context, db pg.DB, logger *log.Logger, txUUID string) (int, error) {
	model := &models.Transaction{}
	err := db.ModelContext(ctx, model).Column("id").Where("uuid = ?", txUUID).Select()
	if err != nil {
		errMsg := "failed to find transaction by uuid"
		logger.WithContext(ctx).WithError(err).Error(errMsg)
		return 0, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ID, nil
}
