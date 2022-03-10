package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/gofrs/uuid"

	"github.com/consensys/orchestrate/src/api/store/models"
)

type PGSchedule struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.ScheduleAgent = &PGSchedule{}

func NewPGSchedule(client postgres.Client) *PGSchedule {
	return &PGSchedule{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.schedule"),
	}
}

func (agent *PGSchedule) Insert(ctx context.Context, schedule *entities.Schedule) error {
	model := models.NewSchedule(schedule)
	model.UUID = uuid.Must(uuid.NewV4()).String()
	model.CreatedAt = time.Now().UTC()

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMessage := "failed to insert schedule"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	utils.CopyPtr(model.ToEntity(), schedule)
	return nil
}

func (agent *PGSchedule) FindOneByUUID(ctx context.Context, scheduleUUID string, tenants []string, ownerID string) (*entities.Schedule, error) {
	schedule := &models.Schedule{}

	err := agent.client.ModelContext(ctx, schedule).
		Relation("Jobs", func(query postgres.Query) (postgres.Query, error) {
			return query.Order("id ASC"), nil
		}).
		Where("schedule.uuid = ?", scheduleUUID).
		WhereAllowedTenants("schedule.tenant_id", tenants).
		WhereAllowedOwner("", ownerID).
		SelectOne()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("schedule not found")
		}

		errMessage := "failed to find schedule"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return schedule.ToEntity(), nil
}

func (agent *PGSchedule) FindAll(ctx context.Context, tenants []string, ownerID string) ([]*entities.Schedule, error) {
	var schedules []models.Schedule

	err := agent.client.ModelContext(ctx, &schedules).
		Relation("Jobs", func(q postgres.Query) (postgres.Query, error) {
			return q.Order("id ASC"), nil
		}).
		WhereAllowedTenants("schedule.tenant_id", tenants).
		WhereAllowedOwner("", ownerID).Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMessage := "failed to find all schedules"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return models.NewSchedules(schedules), nil
}

func getScheduleIDByUUID(ctx context.Context, client postgres.Client, logger *log.Logger, scheduleUUID string) (int, error) {
	model := &models.Schedule{}
	err := client.ModelContext(ctx, model).Column("id").Where("uuid = ?", scheduleUUID).Select()
	if err != nil {
		errMsg := "failed to find schedule by uuid"
		logger.WithContext(ctx).WithError(err).Error(errMsg)
		return 0, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ID, nil
}
