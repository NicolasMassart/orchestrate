// +build unit

package faucets

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteFaucet_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	faucetAgent := mocks.NewMockFaucetAgent(ctrl)
	mockDB.EXPECT().Faucet().Return(faucetAgent).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewDeleteFaucetUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		faucetModel := testdata.FakeFaucet()

		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), "uuid", userInfo.AllowedTenants).Return(faucetModel, nil)
		faucetAgent.EXPECT().Delete(gomock.Any(), faucetModel, userInfo.AllowedTenants).Return(nil)

		err := usecase.Execute(ctx, "uuid", userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if findOne faucet fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), "uuid", userInfo.AllowedTenants).Return(nil, expectedErr)

		err := usecase.Execute(ctx, "uuid", userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteFaucetComponent), err)
	})

	t.Run("should fail with same error if delete faucet fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), "uuid", userInfo.AllowedTenants).Return(testdata.FakeFaucet(), nil)
		faucetAgent.EXPECT().Delete(gomock.Any(), gomock.Any(), userInfo.AllowedTenants).Return(expectedErr)

		err := usecase.Execute(ctx, "uuid", userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteFaucetComponent), err)
	})
}
