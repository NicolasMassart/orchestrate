package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-pg/pg/v10"
	"github.com/gofrs/uuid"
)

type PGSubscription struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.SubscriptionAgent = &PGSubscription{}

func NewPGSubscription(client postgres.Client) *PGSubscription {
	return &PGSubscription{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.subscription"),
	}
}

func (agent *PGSubscription) Insert(ctx context.Context, subscription *entities.Subscription) (*entities.Subscription, error) {
	model := models.NewSubscription(subscription)

	model.UUID = uuid.Must(uuid.NewV4()).String()
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMsg := "failed to insert subscription"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ToEntity(), nil
}

func (agent *PGSubscription) Search(ctx context.Context, filters *entities.SubscriptionFilters, tenants []string, ownerID string) ([]*entities.Subscription, error) {
	var subscriptions []*models.Subscription

	q := agent.client.ModelContext(ctx, &subscriptions)
	if len(filters.Addresses) > 0 {
		addrs := []string{}
		for _, addr := range filters.Addresses {
			addrs = append(addrs, addr.String())
		}
		q = q.Where("address in (?)", pg.In(addrs))
	}
	if filters.TenantID != "" {
		q = q.Where("tenant_id = ?", filters.TenantID)
	}
	if filters.ChainUUID != "" {
		q = q.Where("chain_uuid = ?", filters.ChainUUID)
	}

	err := q.WhereAllowedTenants("", tenants).WhereAllowedOwner("", ownerID).Order("id ASC").Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMsg := "failed to search subscriptions"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return models.NewSubscriptions(subscriptions), nil
}

func (agent *PGSubscription) FindOneByAddressAndTenant(ctx context.Context, address *common.Address, tenantID string, tenants []string, ownerID string) (*entities.Subscription, error) {
	subscription := &models.Subscription{}

	// First, we search for an subscription for the specified chain
	err := agent.client.
		ModelContext(ctx, subscription).
		Where("tenant_id = ?", tenantID).
		Where("address = ?", address.String()).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil && errors.IsNotFoundError(err) {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}
		errMsg := "failed to find one subscription by tenant and address"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return subscription.ToEntity(), nil
}

func (agent *PGSubscription) FindOneByUUID(ctx context.Context, subscriptionUUID string, tenants []string, ownerID string) (*entities.Subscription, error) {
	model := &models.Subscription{}
	err := agent.client.ModelContext(ctx, model).
		Where("uuid = ?", subscriptionUUID).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("subscription not found")
		}

		errMessage := "failed to select subscription"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGSubscription) Update(ctx context.Context, subscription *entities.Subscription, tenants []string, ownerID string) (*entities.Subscription, error) {
	model := models.NewSubscription(subscription)
	model.UpdatedAt = time.Now().UTC()

	err := agent.client.ModelContext(ctx, model).
		Where("uuid = ?", subscription.UUID).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		UpdateNotZero()
	if err != nil {
		errMessage := "failed to update subscription"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}

func (agent *PGSubscription) Delete(ctx context.Context, subscriptionUUID string, tenants []string, ownerID string) error {
	err := agent.client.ModelContext(ctx, &models.Subscription{}).
		Where("uuid = ?", subscriptionUUID).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		Delete()
	if err != nil {
		errMessage := "failed to delete subscription"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	return nil
}
