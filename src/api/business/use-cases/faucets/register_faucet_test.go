// +build unit

package faucets

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/api/store/parsers"
	"github.com/consensys/orchestrate/src/api/store/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterFaucet_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	faucet := testdata.FakeFaucet()
	mockDB := mocks.NewMockDB(ctrl)
	faucetAgent := mocks.NewMockFaucetAgent(ctrl)
	mockSearchFaucetsUC := mocks2.NewMockSearchFaucetsUseCase(ctrl)

	mockDB.EXPECT().Faucet().Return(faucetAgent).AnyTimes()
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewRegisterFaucetUseCase(mockDB, mockSearchFaucetsUC)

	t.Run("should execute use case successfully", func(t *testing.T) {
		faucetModel := parsers.NewFaucetModelFromEntity(faucet)
		faucetModel.TenantID = userInfo.TenantID

		mockSearchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name}, TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Faucet{}, nil)
		faucetAgent.EXPECT().Insert(gomock.Any(), faucetModel).Return(nil)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, parsers.NewFaucetFromModel(faucetModel), resp)
	})

	t.Run("should fail with AlreadyExistsError if search faucets returns results", func(t *testing.T) {
		mockSearchFaucetsUC.EXPECT().
			Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name}, TenantID: userInfo.TenantID},
				userInfo).Return([]*entities.Faucet{faucet}, nil)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		assert.True(t, errors.IsAlreadyExistsError(err))
	})

	t.Run("should fail with same error if search faucets fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockSearchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name}, TenantID: userInfo.TenantID},
			userInfo).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(registerFaucetComponent), err)
	})

	t.Run("should fail with same error if insert faucet fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockSearchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name}, TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Faucet{}, nil)
		faucetAgent.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(registerFaucetComponent), err)
	})
}
