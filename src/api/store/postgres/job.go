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

func (agent *PGJob) Insert(ctx context.Context, job *entities.Job, jobLog *entities.Log) error {
	jobModel := models.NewJob(job)
	if jobModel.UUID == "" {
		jobModel.UUID = uuid.Must(uuid.NewV4()).String()
	}
	jobModel.CreatedAt = time.Now().UTC()
	jobModel.UpdatedAt = jobModel.CreatedAt

	scheduleID, err := getScheduleIDByUUID(ctx, agent.client, job.ScheduleUUID, agent.logger)
	if err != nil {
		return err
	}

	err = agent.client.RunInTransaction(ctx, func(dbtx postgres.Client) error {
		jobModel.ScheduleID = &scheduleID

		txModel := models.NewTransaction(job.Transaction)
		txModel.UUID = uuid.Must(uuid.NewV4()).String()
		txModel.CreatedAt = jobModel.CreatedAt
		txModel.UpdatedAt = jobModel.CreatedAt
		err = agent.client.ModelContext(ctx, txModel).Insert()
		if err != nil {
			return err
		}
		jobModel.TransactionID = &txModel.ID

		err = agent.client.ModelContext(ctx, jobModel).Insert()
		if err != nil {
			return err
		}

		jobLogModel := models.NewLog(jobLog)
		jobLogModel.JobID = &jobModel.ID
		jobLogModel.UUID = uuid.Must(uuid.NewV4()).String()
		jobLogModel.CreatedAt = jobModel.CreatedAt
		err = dbtx.ModelContext(ctx, jobLogModel).Insert()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		errMsg := "failed to insert job"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).SetMessage(errMsg)
	}

	utils.CopyPtr(jobModel.ToEntity(), job)
	return nil
}

func (agent *PGJob) Update(ctx context.Context, job *entities.Job, jobLog *entities.Log) error {
	curJobModel, err := getJobModelUUID(ctx, agent.client, job.UUID, agent.logger)
	if err != nil {
		return err
	}

	jobModel := models.NewJob(job)
	jobModel.UpdatedAt = time.Now().UTC()

	err = agent.client.RunInTransaction(ctx, func(dbtx postgres.Client) error {
		err = agent.client.ModelContext(ctx, jobModel).Where("id = ?", curJobModel.ID).UpdateNotZero()
		if err != nil {
			return err
		}

		if jobLog != nil {
			jobLogModel := models.NewLog(jobLog)
			jobLogModel.JobID = &curJobModel.ID
			jobLogModel.UUID = uuid.Must(uuid.NewV4()).String()
			jobLogModel.CreatedAt = jobModel.UpdatedAt
			err = dbtx.ModelContext(ctx, jobLogModel).Insert()
			if err != nil {
				return err
			}
		}

		if job.Transaction != nil {
			jobTxModel := models.NewTransaction(job.Transaction)
			jobTxModel.UpdatedAt = jobModel.UpdatedAt
			err = dbtx.ModelContext(ctx, jobTxModel).Where("id = ?", *curJobModel.TransactionID).UpdateNotZero()
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		errMsg := "failed to update job"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return errors.FromError(err).SetMessage(errMsg)
	}

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

func (agent *PGJob) GetSiblingJobs(ctx context.Context, parentJobUUID string, tenants []string, ownerID string) ([]*entities.Job, error) {
	var jobs []*models.Job

	q := agent.client.ModelContext(ctx, &jobs).Relation("Schedule")

	q = q.Where(fmt.Sprintf("(%s) OR (%s)",
		"job.internal_data @> '{\"parentJobUUID\": \"?\"}'",
		"job.uuid = '?'",
	), pg.Safe(parentJobUUID), pg.Safe(parentJobUUID))

	err := q.WhereAllowedTenants("schedule.tenant_id", tenants).
		Order("id ASC").
		WhereAllowedOwner("schedule.owner_id", ownerID).
		Select()

	if err != nil {
		errMessage := "failed to find sibling jobs"
		agent.logger.WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return models.NewJobs(jobs), nil
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

func getJobModelUUID(ctx context.Context, client postgres.Client, jobUUID string, logger *log.Logger) (*models.Job, error) {
	model := &models.Job{}
	err := client.ModelContext(ctx, model).Where("uuid = ?", jobUUID).Select()
	if err != nil {
		errMsg := "failed to retrieve job model"
		logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model, nil
}
