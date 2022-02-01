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
	mockTxDA := mocks.NewMockTransactionAgent(ctrl)
	mockStartJobUC := mocks2.NewMockStartJobUseCase(ctrl)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDB.EXPECT().Transaction().Return(mockTxDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewStartNextJobUseCase(mockDB, mockStartJobUC)

	ctx := context.Background()

	t.Run("should execute use case for EEA marking transaction successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		nextJob := testdata.FakeJob()
		txHash := ethcommon.HexToHash("0x123")

		job.NextJobUUID = nextJob.UUID
		job.Transaction.Hash = &txHash
		job.Status = entities.StatusStored
		job.Logs = append(job.Logs, &entities.Log{
			Status:    entities.StatusStored,
			CreatedAt: time.Now(),
		})
		job.Type = entities.EEAPrivateTransaction
		nextJob.Type = entities.EEAMarkingTransaction

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(job, nil)
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(nextJob, nil)
		mockTxDA.EXPECT().Update(gomock.Any(), nextJob.Transaction, nextJob.UUID).Return(nil)

		mockStartJobUC.EXPECT().Execute(gomock.Any(), nextJob.UUID, userInfo)
		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should execute use case for tessera marking transaction successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		nextJob := testdata.FakeJob()
		enclaveKey := ethcommon.HexToHash("0x123")
	
		job.NextJobUUID = nextJob.UUID
		job.Transaction.EnclaveKey = enclaveKey.Bytes()
		job.Transaction.Gas = utils.ToPtr(uint64(11)).(*uint64)
		job.Status = entities.StatusStored
		job.Logs = append(job.Logs, &entities.Log{
			Status:    entities.StatusStored,
			CreatedAt: time.Now(),
		})
		job.Type = entities.TesseraPrivateTransaction
		nextJob.Type = entities.TesseraMarkingTransaction

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(job, nil)
		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), nextJob.UUID, userInfo.AllowedTenants, userInfo.Username, false).
			Return(nextJob, nil)
		nextJob.Transaction.Data = enclaveKey.Bytes()
		nextJob.Transaction.Gas =  utils.ToPtr(uint64(11)).(*uint64)
		mockTxDA.EXPECT().Update(gomock.Any(), nextJob.Transaction, nextJob.UUID).Return(nil)

		mockStartJobUC.EXPECT().Execute(gomock.Any(), nextJob.UUID, userInfo)
		err := usecase.Execute(ctx, job.UUID, userInfo)

		assert.NoError(t, err)
	})
}
