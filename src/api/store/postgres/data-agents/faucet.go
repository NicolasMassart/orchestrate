package dataagents

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"
	gopg "github.com/go-pg/pg/v9"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/gofrs/uuid"
)

const faucetDAComponent = "data-agents.faucet"

// PGFaucet is a Faucet data agent for PostgreSQL
type PGFaucet struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGFaucet creates a new PGFaucet
func NewPGFaucet(db pg.DB) store.FaucetAgent {
	return &PGFaucet{db: db, logger: log.NewLogger().SetComponent(faucetDAComponent)}
}

// Insert Inserts a new faucet in DB
func (agent *PGFaucet) Insert(ctx context.Context, faucet *entities.Faucet) error {
	model := parsers.NewFaucetModel(faucet)
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = time.Now().UTC()
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}

	err := pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert faucet")
		return errors.FromError(err).ExtendComponent(faucetDAComponent)
	}

	utils.CopyPtr(parsers.NewFaucetEntity(model), faucet)
	return nil
}

// FindOneByUUID Finds a faucet in DB
func (agent *PGFaucet) FindOneByUUID(ctx context.Context, faucetUUID string, tenants []string) (*entities.Faucet, error) {
	model := &models.Faucet{}
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", faucetUUID)
	query = pg.WhereAllowedTenants(query, "tenant_id", tenants)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to select faucet")
		}
		return nil, errors.FromError(err).ExtendComponent(faucetDAComponent)
	}

	return parsers.NewFaucetEntity(model), nil
}

func (agent *PGFaucet) Search(ctx context.Context, filters *entities.FaucetFilters, tenants []string) ([]*entities.Faucet, error) {
	var faucets []*models.Faucet

	query := agent.db.ModelContext(ctx, &faucets)
	if len(filters.Names) > 0 {
		query = query.Where("name in (?)", gopg.In(filters.Names))
	}
	if filters.TenantID != "" {
		query = query.Where("tenant_id = ?", filters.TenantID)
	}
	if filters.ChainRule != "" {
		query = query.Where("chain_rule = ?", filters.ChainRule)
	}

	query = pg.WhereAllowedTenants(query, "tenant_id", tenants).Order("created_at ASC")

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to search faucet")
		}
		return nil, errors.FromError(err).ExtendComponent(faucetDAComponent)
	}

	return parsers.NewFaucetEntityArr(faucets), nil
}

func (agent *PGFaucet) Update(ctx context.Context, faucet *entities.Faucet, tenants []string) error {
	model := parsers.NewFaucetModel(faucet)
	model.UpdatedAt = time.Now().UTC()
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", faucet.UUID)
	query = pg.WhereAllowedTenantsDefault(query, tenants)

	err := pg.Update(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to update faucet")
		return errors.FromError(err).ExtendComponent(faucetDAComponent)
	}

	utils.CopyPtr(parsers.NewFaucetEntity(model), faucet)
	return nil
}

func (agent *PGFaucet) Delete(ctx context.Context, faucet *entities.Faucet, tenants []string) error {
	model := parsers.NewFaucetModel(faucet)
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", faucet.UUID)
	query = pg.WhereAllowedTenantsDefault(query, tenants)

	err := pg.Delete(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to delete faucet")
		return errors.FromError(err).ExtendComponent(faucetDAComponent)
	}

	return nil
}
