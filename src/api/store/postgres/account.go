package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/go-pg/pg/v10"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
)

type PGAccount struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.AccountAgent = &PGAccount{}

func NewPGAccount(client postgres.Client) *PGAccount {
	return &PGAccount{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.account"),
	}
}

func (agent *PGAccount) Insert(ctx context.Context, account *entities.Account) (*entities.Account, error) {
	model := models.NewAccount(account)
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = time.Now().UTC()

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMsg := "failed to insert account"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ToEntity(), nil
}

func (agent *PGAccount) Update(ctx context.Context, account *entities.Account) (*entities.Account, error) {
	model := models.NewAccount(account)
	model.UpdatedAt = time.Now().UTC()

	q := agent.client.ModelContext(ctx, model).
		Where("address = ?", account.Address.Hex()).
		Where("tenant_id = ?", account.TenantID)

	if account.OwnerID != "" {
		q = q.Where("owner_id = ?", account.OwnerID)
	}

	err := q.UpdateNotZero()
	if err != nil {
		errMsg := "failed to update account"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ToEntity(), nil
}

func (agent *PGAccount) Search(ctx context.Context, filters *entities.AccountFilters, tenants []string, ownerID string) ([]*entities.Account, error) {
	var accounts []*models.Account

	q := agent.client.ModelContext(ctx, &accounts)
	if len(filters.Aliases) > 0 {
		q = q.Where("alias in (?)", pg.In(filters.Aliases))
	}
	if filters.TenantID != "" {
		q = q.Where("tenant_id = ?", filters.TenantID)
	}

	err := q.WhereAllowedTenants("", tenants).WhereAllowedOwner("", ownerID).Order("id ASC").Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMsg := "failed to search accounts"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return models.NewAccounts(accounts), nil
}

func (agent *PGAccount) FindOneByAddress(ctx context.Context, address string, tenants []string, ownerID string) (*entities.Account, error) {
	account := &models.Account{}

	err := agent.client.
		ModelContext(ctx, account).
		Where("address = ?", address).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("account not found")
		}

		errMsg := "failed to find one account by address"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return account.ToEntity(), nil
}
