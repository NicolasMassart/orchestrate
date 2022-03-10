package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/go-pg/pg/v10"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"
)

type PGChain struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.ChainAgent = &PGChain{}

func NewPGChain(client postgres.Client) *PGChain {
	return &PGChain{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.chain"),
	}
}

func (agent *PGChain) Insert(ctx context.Context, chain *entities.Chain) error {
	model := models.NewChain(chain)
	model.UUID = uuid.Must(uuid.NewV4()).String()
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	err := agent.client.RunInTransaction(ctx, func(c postgres.Client) error {
		err := c.ModelContext(ctx, model).Insert()
		if err != nil {
			errMessage := "failed to insert chain"
			agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
			return errors.FromError(err).SetMessage(errMessage)
		}

		return nil
	})
	if err != nil {
		return err
	}

	utils.CopyPtr(model.ToEntity(), chain)
	return nil
}

func (agent *PGChain) FindOneByUUID(ctx context.Context, chainUUID string, tenants []string, ownerID string) (*entities.Chain, error) {
	chain := &models.Chain{}

	err := agent.client.
		ModelContext(ctx, chain).
		Where("uuid = ?", chainUUID).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("chain not found by uuid")
		}

		errMessage := "failed to find chain by uuid"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return chain.ToEntity(), nil
}

func (agent *PGChain) FindOneByName(ctx context.Context, name string, tenants []string, ownerID string) (*entities.Chain, error) {
	chain := &models.Chain{}

	err := agent.client.
		ModelContext(ctx, chain).
		Where("name = ?", name).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("chain not found by name")
		}

		errMessage := "failed to find chain by name"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return chain.ToEntity(), nil
}

func (agent *PGChain) Search(ctx context.Context, filters *entities.ChainFilters, tenants []string, ownerID string) ([]*entities.Chain, error) {
	var chains []*models.Chain

	q := agent.client.ModelContext(ctx, &chains)

	if len(filters.Names) > 0 {
		q = q.Where("name in (?)", pg.In(filters.Names))
	}
	if filters.TenantID != "" {
		q = q.Where("tenant_id = ?", filters.TenantID)
	}

	err := q.WhereAllowedTenants("", tenants).
		Order("created_at ASC").
		WhereAllowedOwner("", ownerID).
		Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMsg := "failed to search chains"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
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

	return models.NewChains(chains), nil
}

func (agent *PGChain) Update(ctx context.Context, nextChain *entities.Chain, tenants []string, ownerID string) error {
	chain, err := agent.FindOneByUUID(ctx, nextChain.UUID, tenants, ownerID)
	if err != nil {
		return err
	}

	model := models.NewChain(nextChain)
	model.UpdatedAt = time.Now().UTC()

	err = agent.client.ModelContext(ctx, model).
		Where("uuid = ?", chain.UUID).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		Update()
	if err != nil {
		errMessage := "failed to update chain"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	utils.CopyPtr(model.ToEntity(), chain)
	return nil
}

func (agent *PGChain) Delete(ctx context.Context, chain *entities.Chain, tenants []string) error {
	model := models.NewChain(chain)

	err := agent.client.ModelContext(ctx, model).
		Where("uuid = ?", chain.UUID).
		WhereAllowedTenants("", tenants).
		Delete()
	if err != nil {
		errMessage := "failed to deleted chain"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	return nil

}
