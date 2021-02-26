// +build unit

package jobs

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/business/parsers"
	"github.com/ConsenSys/orchestrate/services/api/store/mocks"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
)

func TestGetJob_Execute(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)

	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	usecase := NewGetJobUseCase(mockDB)

	tenantID := "tenantID"

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testutils.FakeJobModel(0)
		expectedResponse := parsers.NewJobEntityFromModels(job)

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), job.UUID, []string{tenantID}, true).Return(job, nil)
		jobResponse, err := usecase.Execute(ctx, job.UUID, []string{tenantID})

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, jobResponse)
	})

	t.Run("should fail with same error if FindOneByUUID fails for job", func(t *testing.T) {
		uuid := "uuid"
		expectedErr := errors.NotFoundError("error")

		mockJobDA.EXPECT().FindOneByUUID(gomock.Any(), uuid, []string{tenantID}, true).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, uuid, []string{tenantID})

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createJobComponent), err)
	})
}
