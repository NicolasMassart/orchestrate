package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/go-pg/pg/v10"

	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/gofrs/uuid"
)

type PGFaucet struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.FaucetAgent = &PGFaucet{}

func NewPGFaucet(client postgres.Client) *PGFaucet {
	return &PGFaucet{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.faucet"),
	}
}

func (agent *PGFaucet) Insert(ctx context.Context, faucet *entities.Faucet) (*entities.Faucet, error) {
	model := models.NewFaucet(faucet)
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt
	model.UUID = uuid.Must(uuid.NewV4()).String()

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMessage := "failed to insert faucet"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGFaucet) FindOneByUUID(ctx context.Context, faucetUUID string, tenants []string) (*entities.Faucet, error) {
	model := &models.Faucet{}
	err := agent.client.ModelContext(ctx, model).
		Where("uuid = ?", faucetUUID).
		WhereAllowedTenants("", tenants).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("faucet not found")
		}

		errMessage := "failed to select faucet"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGFaucet) Search(ctx context.Context, filters *entities.FaucetFilters, tenants []string) ([]*entities.Faucet, error) {
	var faucets []*models.Faucet

	q := agent.client.ModelContext(ctx, &faucets)
	if len(filters.Names) > 0 {
		q = q.Where("name in (?)", pg.In(filters.Names))
	}
	if filters.TenantID != "" {
		q = q.Where("tenant_id = ?", filters.TenantID)
	}
	if filters.ChainRule != "" {
		q = q.Where("chain_rule = ?", filters.ChainRule)
	}

	err := q.WhereAllowedTenants("", tenants).
		Order("created_at ASC").
		Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMessage := "failed to search faucet"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return models.NewFaucets(faucets), nil
}

func (agent *PGFaucet) Update(ctx context.Context, faucet *entities.Faucet, tenants []string) (*entities.Faucet, error) {
	model := models.NewFaucet(faucet)
	model.UpdatedAt = time.Now().UTC()

	err := agent.client.ModelContext(ctx, model).
		Where("uuid = ?", faucet.UUID).
		WhereAllowedTenants("", tenants).
		Update()
	if err != nil {
		errMessage := "failed to update faucet"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGFaucet) Delete(ctx context.Context, faucet *entities.Faucet, tenants []string) error {
	model := models.NewFaucet(faucet)

	err := agent.client.ModelContext(ctx, model).
		Where("uuid = ?", faucet.UUID).
		WhereAllowedTenants("", tenants).
		Delete()
	if err != nil {
		errMessage := "failed to delete faucet"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	return nil
}
