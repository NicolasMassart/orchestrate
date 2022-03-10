// +build unit

package jobs

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateChildren_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	status := entities.StatusNeverMined

	mockDB := mocks.NewMockDB(ctrl)
	mockJobDA := mocks.NewMockJobAgent(ctrl)
	mockLogDA := mocks.NewMockLogAgent(ctrl)
	mockDB.EXPECT().Job().Return(mockJobDA).AnyTimes()
	mockDB.EXPECT().Log().Return(mockLogDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateChildrenUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		parentJobUUID := "parentJobUUID"
		jobUUID := "jobUUID"
		log := &entities.Log{
			Status:  status,
			Message: fmt.Sprintf("sibling (or parent) job %s was mined instead", jobUUID),
		}

		jobsToUpdate := []*entities.Job{testdata.FakeJob(), testdata.FakeJob()}
		jobsToUpdate[0].Status = entities.StatusPending
		jobsToUpdate[1].Status = entities.StatusPending

		mockJobDA.EXPECT().Search(gomock.Any(), &entities.JobFilters{ParentJobUUID: parentJobUUID, Status: entities.StatusPending},
			userInfo.AllowedTenants, userInfo.Username).Return(jobsToUpdate, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), jobsToUpdate[0]).Return(nil)
		mockJobDA.EXPECT().Update(gomock.Any(), jobsToUpdate[1]).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), log, jobsToUpdate[0].UUID).Return(log, nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), log, jobsToUpdate[1].UUID).Return(log, nil)

		err := usecase.Execute(ctx, jobUUID, parentJobUUID, status, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should not update status of the jobUUID job", func(t *testing.T) {
		parentJobUUID := "parentJobUUID"
		jobUUID := "jobUUID"
		log := &entities.Log{
			Status:  status,
			Message: fmt.Sprintf("sibling (or parent) job %s was mined instead", jobUUID),
		}

		jobsToUpdate := []*entities.Job{testdata.FakeJob(), testdata.FakeJob()}
		jobsToUpdate[0].UUID = jobUUID
		jobsToUpdate[0].Status = entities.StatusPending
		jobsToUpdate[1].Status = entities.StatusPending

		mockJobDA.EXPECT().Search(gomock.Any(), &entities.JobFilters{ParentJobUUID: parentJobUUID, Status: entities.StatusPending},
			userInfo.AllowedTenants, userInfo.Username).Return(jobsToUpdate, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), log, jobsToUpdate[1].UUID).Return(log, nil)

		err := usecase.Execute(ctx, jobUUID, parentJobUUID, status, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if Search fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")

		mockJobDA.EXPECT().Search(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		err := usecase.Execute(ctx, "jobUUID", "parentJobUUID", status, userInfo)

		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChildrenComponent), err)
	})

	t.Run("should fail with same error if Insert fails", func(t *testing.T) {
		jobsToUpdate := []*entities.Job{testdata.FakeJob()}
		jobsToUpdate[0].Status = entities.StatusPending

		expectedErr := fmt.Errorf("error")

		mockJobDA.EXPECT().Search(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username).Return(jobsToUpdate, nil)
		mockJobDA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		mockLogDA.EXPECT().Insert(gomock.Any(), gomock.Any(), jobsToUpdate[0].UUID).Return(nil, expectedErr)

		err := usecase.Execute(ctx, "jobUUID", "parentJobUUID", status, userInfo)

		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChildrenComponent), err)
	})
}
