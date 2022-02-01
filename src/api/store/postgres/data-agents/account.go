package dataagents

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"

	"github.com/consensys/orchestrate/src/api/store/models"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities"
	gopg "github.com/go-pg/pg/v9"
)

const accountDAComponent = "data-agents.account"

// PGAccount is an Account data agent for PostgreSQL
type PGAccount struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGAccount creates a new PGAccount
func NewPGAccount(db pg.DB) store.AccountAgent {
	return &PGAccount{
		db:     db,
		logger: log.NewLogger().SetComponent(accountDAComponent),
	}
}

func (agent *PGAccount) Insert(ctx context.Context, account *entities.Account) error {
	model := parsers.NewAccountModel(account)
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = time.Now().UTC()
	agent.db.ModelContext(ctx, model)
	err := pg.Insert(ctx, agent.db, model)
	if err != nil {
		errMsg := "failed to insert account"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).SetMessage(errMsg).ExtendComponent(accountDAComponent)
	}

	utils.CopyPtr(parsers.NewAccountEntity(model), account)
	return nil
}

// Insert Inserts a new job in DB
func (agent *PGAccount) Update(ctx context.Context, account *entities.Account) error {
	model := parsers.NewAccountModel(account)
	model.UpdatedAt = time.Now().UTC()
	query := agent.db.ModelContext(ctx, model).
		Where("address = ?", account.Address.String()).
		Where("tenant_id = ?", account.TenantID)
	if account.OwnerID != "" {
		query = query.Where("owner_id = ?", account.OwnerID)
	}
	err := pg.Update(ctx, query)
	if err != nil {
		errMsg := "failed to update account"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).ExtendComponent(accountDAComponent)
	}

	utils.CopyPtr(parsers.NewAccountEntity(model), account)
	return nil
}

func (agent *PGAccount) Search(ctx context.Context, filters *entities.AccountFilters, tenants []string, ownerID string) ([]*entities.Account, error) {
	var accounts []*models.Account

	query := agent.db.ModelContext(ctx, &accounts)
	if len(filters.Aliases) > 0 {
		query = query.Where("alias in (?)", gopg.In(filters.Aliases))
	}
	if filters.TenantID != "" {
		query = query.Where("tenant_id = ?", filters.TenantID)
	}

	query = pg.WhereAllowedTenants(query, "tenant_id", tenants).Order("id ASC")
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	err := pg.Select(ctx, query)
	if err != nil {
		errMsg := "failed to search accounts"
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		}
		return nil, errors.FromError(err).SetMessage(errMsg).ExtendComponent(accountDAComponent)
	}

	return parsers.NewAccountEntityArr(accounts), nil
}

func (agent *PGAccount) FindOneByAddress(ctx context.Context, address string, tenants []string, ownerID string) (*entities.Account, error) {
	account := &models.Account{}

	query := agent.db.ModelContext(ctx, account).Where("address = ?", address)

	query = pg.WhereAllowedTenants(query, "tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	err := pg.SelectOne(ctx, query)
	if err != nil {
		errMsg := "failed to find one account by address"
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		}
		return nil, errors.FromError(err).SetMessage(errMsg).ExtendComponent(accountDAComponent)
	}

	return parsers.NewAccountEntity(account), nil
}
