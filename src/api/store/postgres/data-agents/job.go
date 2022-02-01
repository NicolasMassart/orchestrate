package dataagents

import (
	"context"
	"fmt"
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
	"github.com/go-pg/pg/v9/orm"
	"github.com/gofrs/uuid"
)

const jobDAComponent = "data-agents.job"

// PGJob is a job data agent for PostgreSQL
type PGJob struct {
	db     pg.DB
	logger *log.Logger
}

// NewPGJob creates a new pgJob
func NewPGJob(db pg.DB) store.JobAgent {
	return &PGJob{db: db, logger: log.NewLogger().SetComponent(jobDAComponent)}
}

// Insert Inserts a new job in DB
func (agent *PGJob) Insert(ctx context.Context, job *entities.Job, scheduleUUID, txUUID string) error {
	model := parsers.NewJobModelFromEntities(job)
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = time.Now().UTC()

	scheduleID, err := getScheduleIDByUUID(ctx, agent.db, agent.logger, scheduleUUID)
	if err != nil {
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}
	model.ScheduleID = &scheduleID

	txID, err := getTxIDByUUID(ctx, agent.db, agent.logger, txUUID)
	if err != nil {
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}
	model.TransactionID = &txID

	agent.db.ModelContext(ctx, model)
	err = pg.Insert(ctx, agent.db, model)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to insert job")
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}

	utils.CopyPtr(parsers.NewJobEntity(model), job)
	return nil
}

// Insert Inserts a new job in DB
func (agent *PGJob) Update(ctx context.Context, job *entities.Job) error {
	model := parsers.NewJobModelFromEntities(job)
	model.UpdatedAt = time.Now().UTC()
	query := agent.db.ModelContext(ctx, model).Where("uuid = ?", job.UUID)
	err := pg.Update(ctx, query)
	if err != nil {
		agent.logger.WithContext(ctx).WithError(err).Error("failed to update job")
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}

	utils.CopyPtr(parsers.NewJobEntity(model), job)
	return nil
}

func (agent *PGJob) FindOneByUUID(ctx context.Context, jobUUID string, tenants []string, ownerID string, withLogs bool) (*entities.Job, error) {
	model, err := agent.findOneByUUID(ctx, jobUUID, tenants, ownerID, withLogs)
	if err != nil {
		return nil, err
	}
	return parsers.NewJobEntity(model), nil
}

// LockOneByUUID gets a job by UUID
func (agent *PGJob) LockOneByUUID(ctx context.Context, jobUUID string) error {
	query := agent.db.ModelContext(ctx, &models.Job{}).Where("job.uuid = ?", jobUUID).For("UPDATE")
	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithError(err).Error("failed to lock job by uuid")
		}
		return errors.FromError(err).ExtendComponent(jobDAComponent)
	}

	return nil
}

func (agent *PGJob) Search(ctx context.Context, filters *entities.JobFilters, tenants []string, ownerID string) ([]*entities.Job, error) {
	var jobs []*models.Job

	query := agent.db.ModelContext(ctx, &jobs).
		Relation("Transaction").
		Relation("Schedule")

	if filters.WithLogs {
		query = query.Relation("Logs", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("id ASC"), nil
		})
	}

	if len(filters.TxHashes) > 0 {
		query = query.Where("transaction.hash in (?)", gopg.In(filters.TxHashes))
	}

	if filters.ChainUUID != "" {
		query = query.Where("job.chain_uuid = ?", filters.ChainUUID)
	}

	if filters.Status != "" {
		query = query.Where("job.status = ?", filters.Status)
	}

	if filters.ParentJobUUID != "" {
		query = query.Where(fmt.Sprintf("(%s) OR (%s)",
			fmt.Sprintf("job.is_parent is false AND job.internal_data @> '{\"parentJobUUID\": \"%s\"}'", filters.ParentJobUUID),
			fmt.Sprintf("job.is_parent is true AND job.uuid = '%s'", filters.ParentJobUUID),
		))
	}

	if filters.OnlyParents {
		query = query.Where("job.is_parent is true")
	}

	if filters.UpdatedAfter.Second() > 0 {
		query = query.Where("job.updated_at >= ?", filters.UpdatedAfter)
	}

	query = pg.WhereAllowedTenants(query, "schedule.tenant_id", tenants).Order("id ASC")
	query = pg.WhereAllowedOwner(query, "schedule.owner_id", ownerID)

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithError(err).Error("failed to search jobs")
		}
		return nil, errors.FromError(err).ExtendComponent(jobDAComponent)
	}

	return parsers.NewJobEntityArr(jobs), nil
}

func (agent *PGJob) findOneByUUID(ctx context.Context, jobUUID string, tenants []string, ownerID string, withLogs bool) (*models.Job, error) {
	job := &models.Job{}

	query := agent.db.ModelContext(ctx, job).
		Where("job.uuid = ?", jobUUID).
		Relation("Transaction").
		Relation("Schedule")

	if withLogs {
		query = query.Relation("Logs", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("id ASC"), nil
		})
	}

	query = pg.WhereAllowedTenants(query, "schedule.tenant_id", tenants).Order("id ASC")
	query = pg.WhereAllowedOwner(query, "schedule.owner_id", ownerID)

	err := pg.Select(ctx, query)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			agent.logger.WithContext(ctx).WithError(err).Error("failed to find job by uuid")
		}
		return nil, errors.FromError(err).ExtendComponent(jobDAComponent)
	}

	return job, nil
}

func getJobIDByUUID(ctx context.Context, db pg.DB, logger *log.Logger, jobUUID string) (int, error) {
	model := &models.Job{}
	err := db.ModelContext(ctx, model).Column("id").Where("uuid = ?", jobUUID).Select()
	if err != nil {
		errMsg := "failed to find job by uuid"
		logger.WithContext(ctx).WithError(err).Error(errMsg)
		return 0, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ID, nil
}
