// +build unit

package faucets

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFaucet_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	faucetAgent := mocks.NewMockFaucetAgent(ctrl)
	mockDB.EXPECT().Faucet().Return(faucetAgent).AnyTimes()
	faucet := testdata.FakeFaucet()
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateFaucetUseCase(mockDB)

	faucet.TenantID = userInfo.TenantID
	t.Run("should execute use case successfully", func(t *testing.T) {
		faucetAgent.EXPECT().Update(gomock.Any(), faucet, userInfo.AllowedTenants).Return(nil)
		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.UUID, userInfo.AllowedTenants).Return(faucet, nil)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, faucet, resp)
	})

	t.Run("should fail with same error if update faucet fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		faucetAgent.EXPECT().Update(gomock.Any(), gomock.Any(), userInfo.AllowedTenants).Return(expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateFaucetComponent), err)
	})
}
