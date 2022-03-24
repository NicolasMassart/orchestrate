package jobs

import (
	"context"
	"fmt"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const startNextJobComponent = "use-cases.next-job-start"

type startNextJobUseCase struct {
	db              store.DB
	startJobUseCase usecases.StartJobUseCase
	logger          *log.Logger
}

func NewStartNextJobUseCase(db store.DB, startJobUC usecases.StartJobUseCase) usecases.StartNextJobUseCase {
	return &startNextJobUseCase{
		db:              db,
		startJobUseCase: startJobUC,
		logger:          log.NewLogger().SetComponent(startNextJobComponent),
	}
}

// Execute gets a job
func (uc *startNextJobUseCase) Execute(ctx context.Context, jobUUID string, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("job", jobUUID))
	logger := uc.logger.WithContext(ctx).WithField("job", jobUUID)
	logger.Debug("starting job")
	job, err := uc.db.Job().FindOneByUUID(ctx, jobUUID, userInfo.AllowedTenants, userInfo.Username, false)
	if err != nil {
		return errors.FromError(err).ExtendComponent(startNextJobComponent)
	}

	if job.NextJobUUID == "" {
		errMsg := fmt.Sprintf("job %s does not have a next job to start", job.NextJobUUID)
		logger.Error(errMsg)
		return errors.DataError(errMsg)
	}

	logger = logger.WithField("next_job", job.NextJobUUID)
	logger.Debug("start next job use-case")

	nextJob, err := uc.db.Job().FindOneByUUID(ctx, job.NextJobUUID, userInfo.AllowedTenants, userInfo.Username, false)
	if err != nil {
		return errors.FromError(err).ExtendComponent(startNextJobComponent)
	}

	switch nextJob.Type {
	case entities.EEAMarkingTransaction:
		err = uc.handleEEAMarkingTx(ctx, job, nextJob)
	case entities.TesseraMarkingTransaction:
		err = uc.handleTesseraMarkingTx(ctx, job, nextJob)
	}

	if err != nil {
		logger.WithError(err).Error("failed to validate next transaction data")
		return errors.FromError(err).ExtendComponent(startNextJobComponent)
	}

	return uc.startJobUseCase.Execute(ctx, nextJob.UUID, userInfo)
}

func (uc *startNextJobUseCase) handleEEAMarkingTx(ctx context.Context, prevJob, job *entities.Job) error {
	if prevJob.Type != entities.EEAPrivateTransaction {
		return errors.DataError("expected previous job as type: %s", entities.EEAPrivateTransaction)
	}

	if prevJob.Status != entities.StatusStored {
		return errors.DataError("expected previous job status as: STORED")
	}

	job.Transaction.Data = prevJob.Transaction.Hash.Bytes()
	return uc.db.Job().Update(ctx, job, nil)
}

func (uc *startNextJobUseCase) handleTesseraMarkingTx(ctx context.Context, prevJob, job *entities.Job) error {
	if prevJob.Type != entities.TesseraPrivateTransaction {
		return errors.DataError("expected previous job as type: %s", entities.TesseraPrivateTransaction)
	}

	if prevJob.Status != entities.StatusStored {
		return errors.DataError("expected previous job status as: STORED")
	}

	job.Transaction.Data = prevJob.Transaction.EnclaveKey
	if prevJob.Transaction.Gas != nil && *prevJob.Transaction.Gas < uint64(entities.TesseraGasLimit) {
		job.Transaction.Gas = utils.ToPtr(uint64(entities.TesseraGasLimit)).(*uint64)
	} else {
		job.Transaction.Gas = prevJob.Transaction.Gas
	}

	err := uc.db.Job().Update(ctx, job, nil)
	if err != nil {
		return err
	}

	return nil
}
