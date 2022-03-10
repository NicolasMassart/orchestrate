package jobs

import (
	"context"
	"math/big"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/store"
)

const createJobComponent = "use-cases.create-job"

type createJobUseCase struct {
	db             store.DB
	getChainUC     usecases.GetChainUseCase
	logger         *log.Logger
	defaultStoreID string
}

func NewCreateJobUseCase(db store.DB, getChainUC usecases.GetChainUseCase, qkmStoreID string) usecases.CreateJobUseCase {
	return &createJobUseCase{
		db:             db,
		getChainUC:     getChainUC,
		logger:         log.NewLogger().SetComponent(createJobComponent),
		defaultStoreID: qkmStoreID,
	}
}

func (uc createJobUseCase) WithDBTransaction(dbtx store.DB) usecases.CreateJobUseCase {
	uc.db = dbtx
	return &uc
}

func (uc *createJobUseCase) Execute(ctx context.Context, job *entities.Job, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	ctx = log.WithFields(ctx, log.Field("chain", job.ChainUUID), log.Field("schedule", job.ScheduleUUID))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("creating new job")

	chainID, err := uc.getChainID(ctx, job.ChainUUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createJobComponent)
	}
	job.InternalData.ChainID = chainID

	if job.Transaction.From != nil && job.Type != entities.EthereumRawTransaction {
		job.InternalData.StoreID, err = uc.getAccountStoreID(ctx, job.Transaction.From, userInfo)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(createJobComponent)
		}
	}

	_, err = uc.db.Schedule().FindOneByUUID(ctx, job.ScheduleUUID, userInfo.AllowedTenants, userInfo.Username)
	if errors.IsNotFoundError(err) {
		return nil, errors.InvalidParameterError("schedule does not exist").ExtendComponent(createJobComponent)
	} else if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createJobComponent)
	}

	job.TenantID = userInfo.TenantID
	job.OwnerID = userInfo.Username
	job.Status = entities.StatusCreated
	job.Logs = append(job.Logs, &entities.Log{
		Status: entities.StatusCreated,
	})

	job.Transaction, err = uc.db.Transaction().Insert(ctx, job.Transaction)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(createJobComponent)
	}

	err = uc.db.RunInTransaction(ctx, func(dbtx store.DB) error {
		// If it's a child job, only create it if parent status is PENDING
		if job.InternalData.ParentJobUUID != "" {
			parentJobUUID := job.InternalData.ParentJobUUID
			logger.WithField("parent_job", parentJobUUID).Debug("lock parent job row for update")
			if der := dbtx.Job().LockOneByUUID(ctx, job.InternalData.ParentJobUUID); der != nil {
				return der
			}

			parentJobModel, der := dbtx.Job().FindOneByUUID(ctx, parentJobUUID, userInfo.AllowedTenants, userInfo.Username, false)
			if der != nil {
				return der
			}

			if parentJobModel.Status != entities.StatusPending {
				errMessage := "cannot create a child job in a finalized schedule"
				logger.WithField("parent_job", parentJobUUID).
					WithField("parent_status", parentJobModel.Status).Error(errMessage)
				return errors.InvalidStateError(errMessage)
			}
		}

		der := dbtx.Job().Insert(ctx, job, job.ScheduleUUID, job.Transaction.UUID)
		if der != nil {
			return der
		}

		job.Logs[0], der = dbtx.Log().Insert(ctx, job.Logs[0], job.UUID)
		if der != nil {
			return der
		}

		return nil
	})
	if err != nil {
		logger.WithError(err).Info("failed to create job")
		return nil, errors.FromError(err).ExtendComponent(createJobComponent)
	}

	logger.WithField("job", job.UUID).Info("job created successfully")
	return job, nil
}

// nolint
func (uc *createJobUseCase) getAccountStoreID(ctx context.Context, address *ethcommon.Address, userInfo *multitenancy.UserInfo) (string, error) {
	acc, err := uc.db.Account().FindOneByAddress(ctx, address.String(), userInfo.AllowedTenants, userInfo.Username)
	if errors.IsNotFoundError(err) {
		return "", errors.InvalidParameterError("failed to get account")
	}
	if err != nil {
		return "", err
	}

	if acc.StoreID == "" {
		return uc.defaultStoreID, nil
	}

	return acc.StoreID, nil
}

func (uc *createJobUseCase) getChainID(ctx context.Context, chainUUID string, userInfo *multitenancy.UserInfo) (*big.Int, error) {
	chain, err := uc.getChainUC.Execute(ctx, chainUUID, userInfo)
	if errors.IsNotFoundError(err) {
		return nil, errors.InvalidParameterError("failed to get chain")
	}
	if err != nil {
		return nil, errors.FromError(err)
	}

	return chain.ChainID, nil
}
