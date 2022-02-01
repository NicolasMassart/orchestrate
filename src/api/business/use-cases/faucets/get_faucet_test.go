// +build unit

package faucets

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetFaucet_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	faucetAgent := mocks.NewMockFaucetAgent(ctrl)
	mockDB.EXPECT().Faucet().Return(faucetAgent).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewGetFaucetUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		faucet := testdata.FakeFaucet()
		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.UUID, userInfo.AllowedTenants).Return(faucet, nil)

		resp, err := usecase.Execute(ctx, faucet.UUID, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, faucet, resp)
	})

	t.Run("should fail with same error if get faucet fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), "uuid", userInfo.AllowedTenants).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, "uuid", userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getFaucetCandidateComponent), err)
	})
}
