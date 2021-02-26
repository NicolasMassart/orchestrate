// +build unit

package jobs

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/ConsenSys/orchestrate/pkg/types/entities"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/business/parsers"
	"github.com/ConsenSys/orchestrate/services/api/store/mocks"
	"github.com/ConsenSys/orchestrate/services/api/store/models"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
)

func TestSearchJobs_Execute(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()

	usecase := NewSearchJobsUseCase(mockDB)

	tenantID := "tenantID"
	tenants := []string{tenantID}

	t.Run("should execute use case successfully", func(t *testing.T) {
		txHash := common.HexToHash("0x1")
		jobs := []*models.Job{testutils.FakeJobModel(0)}
		chainUUID := uuid.Must(uuid.NewV4()).String()
		filters := &entities.JobFilters{
			TxHashes:  []string{txHash.String()},
			ChainUUID: chainUUID,
		}

		expectedResponse := []*entities.Job{parsers.NewJobEntityFromModels(jobs[0])}
		mockJobDA.EXPECT().Search(gomock.Any(), filters, tenants).Return(jobs, nil)

		jobResponse, err := usecase.Execute(ctx, filters, tenants)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, jobResponse)
	})

	t.Run("should fail with same error if search fails for jobs", func(t *testing.T) {
		filters := &entities.JobFilters{}
		expectedErr := errors.NotFoundError("error")

		mockJobDA.EXPECT().Search(gomock.Any(), filters, tenants).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, filters, tenants)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchJobsComponent), err)
	})
}
