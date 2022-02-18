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
	gopg "github.com/go-pg/pg/v9"
	"github.com/gofrs/uuid"
)

const chainDAComponent = "data-agents.chain"

// PGChain is a Chain data agent for PostgreSQL
type PGChain struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGChain creates a new PGChain
func NewPGChain(db pg.DB) store.ChainAgent {
	return &PGChain{db: db, logger: log.NewLogger().SetComponent(chainDAComponent)}
}

// Insert Inserts a new chain in DB
func (agent *PGChain) Insert(ctx context.Context, chain *entities.Chain) error {
	model := parsers.NewChainModel(chain)
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}

	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt
	err := pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert chain")
		return errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	utils.CopyPtr(parsers.NewChainEntity(model), chain)
	return nil
}

// FindOneByUUID Finds a chain in DB by UUID
func (agent *PGChain) FindOneByUUID(ctx context.Context, chainUUID string, tenants []string, ownerID string) (*entities.Chain, error) {
	model := &models.Chain{}
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", chainUUID).Relation("PrivateTxManagers")
	query = pg.WhereAllowedTenants(query, "tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to select chain by uuid")
		}
		return nil, errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	return parsers.NewChainEntity(model), nil
}

// FindOneByName Finds a chain in DB by name
func (agent *PGChain) FindOneByName(ctx context.Context, name string, tenants []string, ownerID string) (*entities.Chain, error) {
	chain := &models.Chain{}

	query := agent.db.ModelContext(ctx, chain).Where("name = ?", name)
	query = pg.WhereAllowedTenants(query, "tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to select chain by name")
		}
		return nil, errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	return parsers.NewChainEntity(chain), nil
}

func (agent *PGChain) Search(ctx context.Context, filters *entities.ChainFilters, tenants []string, ownerID string) ([]*entities.Chain, error) {
	var chains []*models.Chain

	query := agent.db.ModelContext(ctx, &chains)

	if len(filters.Names) > 0 {
		query = query.Where("name in (?)", gopg.In(filters.Names))
	}
	if filters.TenantID != "" {
		query = query.Where("tenant_id = ?", filters.TenantID)
	}

	query = pg.WhereAllowedTenants(query, "tenant_id", tenants).Order("created_at ASC")
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	if err := pg.Select(ctx, query); err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to search chains")
		}
		return nil, errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	if len(chains) == 0 {
		return []*entities.Chain{}, nil
	}

	// We manually link chains to privateTxManager avoiding GO-PG strategy of one query per chain
	chainUUIDs := make([]string, len(chains))
	chainsMap := map[string]*models.Chain{}
	for idx, c := range chains {
		chainUUIDs[idx] = c.UUID
		chainsMap[c.UUID] = c
	}

	var privateTxManagers []*models.PrivateTxManager
	query = agent.db.ModelContext(ctx, &privateTxManagers)
	query = query.Where("chain_uuid in (?)", gopg.In(chainUUIDs))

	if err := pg.Select(ctx, query); err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to search for private tx manager")
		return nil, errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	for _, pm := range privateTxManagers {
		chainsMap[pm.ChainUUID].PrivateTxManagers = []*models.PrivateTxManager{pm}
	}

	return parsers.NewChainEntityArr(chains), nil
}

func (agent *PGChain) Update(ctx context.Context, chain *entities.Chain, tenants []string, ownerID string) error {
	model := parsers.NewChainModel(chain)
	model.UpdatedAt = time.Now().UTC()
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", chain.UUID)
	query = pg.WhereAllowedTenantsDefault(query, tenants)
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	err := pg.Update(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to update chain")
		return errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	utils.CopyPtr(parsers.NewChainEntity(model), chain)
	return nil
}

func (agent *PGChain) Delete(ctx context.Context, chain *entities.Chain, tenants []string) error {
	model := parsers.NewChainModel(chain)
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", chain.UUID)
	query = pg.WhereAllowedTenantsDefault(query, tenants)

	err := pg.Delete(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to deleted chain")
		return errors.FromError(err).ExtendComponent(chainDAComponent)
	}

	return nil
}
