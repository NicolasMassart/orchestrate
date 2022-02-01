// +build unit

package jobs

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSearchJobs_Execute(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewSearchJobsUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		txHash := common.HexToHash("0x1")
		jobs := []*entities.Job{testdata.FakeJob()}
		chainUUID := uuid.Must(uuid.NewV4()).String()
		filters := &entities.JobFilters{
			TxHashes:  []string{txHash.String()},
			ChainUUID: chainUUID,
		}

		mockJobDA.EXPECT().Search(gomock.Any(), filters, userInfo.AllowedTenants, userInfo.Username).Return(jobs, nil)

		jobResponse, err := usecase.Execute(ctx, filters, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, jobs, jobResponse)
	})

	t.Run("should fail with same error if search fails for jobs", func(t *testing.T) {
		filters := &entities.JobFilters{}
		expectedErr := errors.NotFoundError("error")

		mockJobDA.EXPECT().Search(gomock.Any(), filters, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, filters, userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchJobsComponent), err)
	})
}
