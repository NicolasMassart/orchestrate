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
	"github.com/go-pg/pg/v10"
	"github.com/gofrs/uuid"
)

type PGEventStream struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.EventStreamAgent = &PGEventStream{}

func NewPGEventStream(client postgres.Client) *PGEventStream {
	return &PGEventStream{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.event_stream"),
	}
}

func (agent *PGEventStream) Insert(ctx context.Context, eventStream *entities.EventStream) (*entities.EventStream, error) {
	model := models.NewEventStream(eventStream)

	model.UUID = uuid.Must(uuid.NewV4()).String()
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMsg := "failed to insert event stream"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ToEntity(), nil
}

func (agent *PGEventStream) Search(ctx context.Context, filters *entities.EventStreamFilters, tenants []string, ownerID string) ([]*entities.EventStream, error) {
	var eventStreams []*models.EventStream

	q := agent.client.ModelContext(ctx, &eventStreams)
	if len(filters.Names) > 0 {
		q = q.Where("name in (?)", pg.In(filters.Names))
	}
	if filters.TenantID != "" {
		q = q.Where("tenant_id = ?", filters.TenantID)
	}

	err := q.WhereAllowedTenants("", tenants).WhereAllowedOwner("", ownerID).Order("id ASC").Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMsg := "failed to search event streams"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return models.NewEventStreams(eventStreams), nil
}

func (agent *PGEventStream) FindOneByTenantAndChain(ctx context.Context, tenantID, chainUUID string, tenants []string, ownerID string) (*entities.EventStream, error) {
	eventStream := &models.EventStream{}

	// First, we check if an event stream for all chains is defined
	err := agent.client.
		ModelContext(ctx, eventStream).
		Where("tenant_id = ?", tenantID).
		WhereAllowedTenants("", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil {
		// If not found, we search for an event stream for the specified chain
		if errors.IsNotFoundError(err) {
			err = agent.client.
				ModelContext(ctx, eventStream).
				Where("tenant_id = ?", tenantID).
				Where("chain_uuid = ?", chainUUID).
				WhereAllowedTenants("", tenants).
				WhereAllowedOwner("", ownerID).
				SelectOne()
			if err != nil {
				if errors.IsNotFoundError(err) {
					return nil, nil
				}
				errMsg := "failed to find one event stream by tenant and chain"
				agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
				return nil, errors.FromError(err).SetMessage(errMsg)
			}
		}
		errMsg := "failed to find one event stream by tenant"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return eventStream.ToEntity(), nil
}
