package dataagents

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/entities"
	pg "github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/go-pg/pg/v9/orm"
	"github.com/gofrs/uuid"

	"github.com/consensys/orchestrate/src/api/store/models"
)

const scheduleDAComponent = "data-agents.schedule"

// PGSchedule is a schedule data agent for PostgreSQL
type PGSchedule struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGSchedule creates a new PGSchedule
func NewPGSchedule(db pg.DB) store.ScheduleAgent {
	return &PGSchedule{db: db, logger: log.NewLogger().SetComponent(scheduleDAComponent)}
}

// Insert Inserts a new schedule in DB
func (agent *PGSchedule) Insert(ctx context.Context, schedule *entities.Schedule) error {
	model := parsers.NewScheduleModel(schedule)
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}
	model.CreatedAt = time.Now().UTC()

	err := pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert schedule")
		return errors.FromError(err).ExtendComponent(scheduleDAComponent)
	}

	utils.CopyPtr(parsers.NewScheduleEntity(model), schedule)
	return nil
}

func (agent *PGSchedule) FindOneByUUID(ctx context.Context, scheduleUUID string, tenants []string, ownerID string) (*entities.Schedule, error) {
	model, err := agent.findOneByUUID(ctx, scheduleUUID, tenants, ownerID)
	if err != nil {
		return nil, err
	}
	return parsers.NewScheduleEntity(model), nil
}

func (agent *PGSchedule) FindAll(ctx context.Context, tenants []string, ownerID string) ([]*entities.Schedule, error) {
	var schedules []*models.Schedule

	query := agent.db.ModelContext(ctx, &schedules).
		Relation("Jobs", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("id ASC"), nil
		})

	query = pg.WhereAllowedTenants(query, "schedule.tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to insert all schedules")
		}
		return nil, errors.FromError(err).ExtendComponent(jobDAComponent)
	}

	return parsers.NewScheduleEntityArr(schedules), nil
}

func (agent *PGSchedule) findOneByUUID(ctx context.Context, scheduleUUID string, tenants []string, ownerID string) (*models.Schedule, error) {
	schedule := &models.Schedule{}

	query := agent.db.ModelContext(ctx, schedule).
		Relation("Jobs", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("id ASC"), nil
		}).
		Where("schedule.uuid = ?", scheduleUUID)

	if ownerID != "" {
		query = query.Where("owner_id = ? OR owner_id IS NULL", ownerID)
	} else {
		query = query.Where("owner_id IS NULL")
	}

	query = pg.WhereAllowedTenants(query, "schedule.tenant_id", tenants)
	query = pg.WhereAllowedOwner(query, "owner_id", ownerID)

	if err := pg.SelectOne(ctx, query); err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to find schedule")
		}
		return nil, errors.FromError(err).ExtendComponent(scheduleDAComponent)
	}

	return schedule, nil
}

func getScheduleIDByUUID(ctx context.Context, db pg.DB, logger *log.Logger, scheduleUUID string) (int, error) {
	model := &models.Schedule{}
	err := db.ModelContext(ctx, model).Column("id").Where("uuid = ?", scheduleUUID).Select()
	if err != nil {
		errMsg := "failed to find schedule by uuid"
		logger.WithContext(ctx).WithError(err).Error(errMsg)
		return 0, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ID, nil
}
