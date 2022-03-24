// +build unit

package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestStartNextJob_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)
	mockStartJobUC := mocks2.NewMockStartJobUseCase(ctrl)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewStartNextJobUseCase(mockDB, mockStartJobUC)

	ctx := context.Background()

	t.Run("should execute use case for EEA marking transaction successfully", func(t *testing.T) {
		prevJob := testdata.FakeJob()
		nextJob := testdata.FakeJob()
		txHash := ethcommon.HexToHash("0x123")

		prevJob.NextJobUUID = nextJob.UUID
		prevJob.Transaction.Hash = &txHash
		prevJob.Status = entities.StatusStored
		prevJob.Logs = append(prevJob.Logs, &entities.Log{
			Status:    entities.StatusStored,
			CreatedAt: time.Now(),
		})
		prevJob.Type = entities.EEAPrivateTransaction
		nextJob.Type = entities.EEAMarkingTransaction

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), prevJob.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(prevJob, nil)
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(nextJob, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, job.Transaction.Data.String(), prevJob.Transaction.Hash.String())
				return nil
			})

		mockStartJobUC.EXPECT().Execute(gomock.Any(), nextJob.UUID, userInfo)
		err := usecase.Execute(ctx, prevJob.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for tessera marking transaction successfully", func(t *testing.T) {
		prevJob := testdata.FakeJob()
		nextJob := testdata.FakeJob()
		enclaveKey := ethcommon.HexToHash("0x123")

		prevJob.NextJobUUID = nextJob.UUID
		prevJob.Transaction.EnclaveKey = enclaveKey.Bytes()
		prevJob.Transaction.Gas = utils.ToPtr(uint64(entities.TesseraGasLimit)).(*uint64)
		prevJob.Status = entities.StatusStored
		prevJob.Logs = append(prevJob.Logs, &entities.Log{
			Status:    entities.StatusStored,
			CreatedAt: time.Now(),
		})
		prevJob.Type = entities.TesseraPrivateTransaction
		nextJob.Type = entities.TesseraMarkingTransaction

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), prevJob.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(prevJob, nil)
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(nextJob, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, job *entities.Job, log *entities.Log) error {
				assert.Equal(t, job.Transaction.Data, prevJob.Transaction.EnclaveKey)
				assert.Equal(t, job.Transaction.Gas, prevJob.Transaction.Gas)
				return nil
			})

		mockStartJobUC.EXPECT().Execute(gomock.Any(), nextJob.UUID, userInfo)
		err := usecase.Execute(ctx, prevJob.UUID, userInfo)

		assert.NoError(t, err)
	})
}
