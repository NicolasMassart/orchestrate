package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/go-pg/pg/v10"
	"github.com/gofrs/uuid"
)

type PGJob struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.JobAgent = &PGJob{}

func NewPGJob(client postgres.Client) *PGJob {
	return &PGJob{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.job"),
	}
}

func (agent *PGJob) Insert(ctx context.Context, job *entities.Job, scheduleUUID, txUUID string) error {
	model := models.NewJob(job)
	if model.UUID == "" {
		model.UUID = uuid.Must(uuid.NewV4()).String()
	}
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	scheduleID, err := getScheduleIDByUUID(ctx, agent.client, agent.logger, scheduleUUID)
	if err != nil {
		return err
	}
	model.ScheduleID = &scheduleID

	tx := &models.Transaction{}
	err = agent.client.ModelContext(ctx, tx).Column("id").Where("uuid = ?", txUUID).Select()
	if err != nil {
		errMsg := "failed to find transaction by uuid"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).SetMessage(errMsg)
	}
	model.TransactionID = &tx.ID

	err = agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMessage := "failed to insert job"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	utils.CopyPtr(model.ToEntity(), job)
	return nil
}

func (agent *PGJob) Update(ctx context.Context, job *entities.Job) error {
	model := models.NewJob(job)
	model.UpdatedAt = time.Now().UTC()

	err := agent.client.ModelContext(ctx, model).Where("uuid = ?", job.UUID).Update()
	if err != nil {
		errMessage := "failed to update job"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	utils.CopyPtr(model.ToEntity(), job)
	return nil
}

func (agent *PGJob) FindOneByUUID(ctx context.Context, jobUUID string, tenants []string, ownerID string, withLogs bool) (*entities.Job, error) {
	job := &models.Job{}

	q := agent.client.ModelContext(ctx, job).
		Where("job.uuid = ?", jobUUID).
		Relation("Transaction").
		Relation("Schedule")

	if withLogs {
		q = q.Relation("Logs", func(query postgres.Query) (postgres.Query, error) {
			return query.Order("id ASC"), nil
		})
	}

	err := q.WhereAllowedTenants("schedule.tenant_id", tenants).
		Order("id ASC").
		WhereAllowedOwner("schedule.owner_id", ownerID).
		Select()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.FromError(err).SetMessage("job not found by uuid")
		}

		errMsg := "failed to find job by uuid"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return job.ToEntity(), nil
}

func (agent *PGJob) LockOneByUUID(ctx context.Context, jobUUID string) error {
	err := agent.client.ModelContext(ctx, &models.Job{}).
		Where("job.uuid = ?", jobUUID).
		For("UPDATE").
		Select()
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil
		}

		errMessage := "failed to lock job by uuid"
		agent.logger.WithError(err).Error(errMessage)
		return errors.FromError(err).SetMessage(errMessage)
	}

	return nil
}

func (agent *PGJob) Search(ctx context.Context, filters *entities.JobFilters, tenants []string, ownerID string) ([]*entities.Job, error) {
	var jobs []*models.Job

	q := agent.client.ModelContext(ctx, &jobs).Relation("Transaction").Relation("Schedule")

	if filters.WithLogs {
		q = q.Relation("Logs", func(q postgres.Query) (postgres.Query, error) {
			return q.Order("id ASC"), nil
		})
	}

	if len(filters.TxHashes) > 0 {
		q = q.Where("transaction.hash in (?)", pg.In(filters.TxHashes))
	}

	if filters.ChainUUID != "" {
		q = q.Where("job.chain_uuid = ?", filters.ChainUUID)
	}

	if filters.Status != "" {
		q = q.Where("job.status = ?", filters.Status)
	}

	if filters.ParentJobUUID != "" {
		q = q.Where(fmt.Sprintf("(%s) OR (%s)",
			"job.is_parent is false AND job.internal_data @> '{\"parentJobUUID\": \"?\"}'",
			"job.is_parent is true AND job.uuid = '?'",
		), pg.Safe(filters.ParentJobUUID), pg.Safe(filters.ParentJobUUID))
	}

	if filters.OnlyParents {
		q = q.Where("job.is_parent is true")
	}

	if filters.UpdatedAfter.Second() > 0 {
		q = q.Where("job.updated_at >= ?", filters.UpdatedAfter)
	}

	err := q.WhereAllowedTenants("schedule.tenant_id", tenants).
		Order("id ASC").
		WhereAllowedOwner("schedule.owner_id", ownerID).
		Select()
	if err != nil && !errors.IsNotFoundError(err) {
		errMessage := "failed to search jobs"
		agent.logger.WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return models.NewJobs(jobs), nil
}
