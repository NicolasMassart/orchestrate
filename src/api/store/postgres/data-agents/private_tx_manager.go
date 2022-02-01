package dataagents

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/gofrs/uuid"
)

const privateTxManagerDAComponent = "data-agents.private-tx-manager"

// PGPrivateTxManager is a Faucet data agent for PostgreSQL
type PGPrivateTxManager struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGPrivateTxManager creates a new PGPrivateTxManager
func NewPGPrivateTxManager(db pg.DB) store.PrivateTxManagerAgent {
	return &PGPrivateTxManager{db: db, logger: log.NewLogger().SetComponent(privateTxManagerDAComponent)}
}

// Insert Inserts a new private transaction manager in DB
func (agent *PGPrivateTxManager) Insert(ctx context.Context, privateTxManager *entities.PrivateTxManager) error {
	if privateTxManager.UUID == "" {
		privateTxManager.UUID = uuid.Must(uuid.NewV4()).String()
	}

	model := parsers.NewPrivateTxManagerModel(privateTxManager)
	err := pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert private tx manager")
		return errors.FromError(err).ExtendComponent(privateTxManagerDAComponent)
	}

	return nil
}

func (agent *PGPrivateTxManager) Search(ctx context.Context, chainUUID string) ([]*entities.PrivateTxManager, error) {
	var privateTxManagers []*models.PrivateTxManager

	query := agent.db.ModelContext(ctx, &privateTxManagers)
	if chainUUID != "" {
		query = query.Where("chain_uuid = ?", chainUUID)
	}

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to search private tx managers")
		}
		return nil, errors.FromError(err).ExtendComponent(privateTxManagerDAComponent)
	}

	return parsers.NewPrivateTxManagerEntityArr(privateTxManagers), nil
}

func (agent *PGPrivateTxManager) Update(ctx context.Context, privateTxManager *entities.PrivateTxManager) error {
	model := parsers.NewPrivateTxManagerModel(privateTxManager)
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", privateTxManager.UUID)

	err := pg.Update(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to update private tx manager")
		return errors.FromError(err).ExtendComponent(privateTxManagerDAComponent)
	}

	utils.CopyPtr(parsers.NewPrivateTxManagerEntity(model), privateTxManager)
	return nil
}

func (agent *PGPrivateTxManager) Delete(ctx context.Context, privateTxManagerUUID string) error {
	query := agent.db.ModelContext(ctx, &models.PrivateTxManager{}).Where("uuid = ?", privateTxManagerUUID)
	err := pg.Delete(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to delete private tx manager")
		return errors.FromError(err).ExtendComponent(privateTxManagerDAComponent)
	}

	return nil
}
