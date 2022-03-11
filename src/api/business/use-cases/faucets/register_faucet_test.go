// +build unit

package faucets

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/stretchr/testify/require"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterFaucet_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	accountAgent := mocks.NewMockAccountAgent(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)
	faucetAgent := mocks.NewMockFaucetAgent(ctrl)

	mockDB.EXPECT().Chain().AnyTimes().Return(chainAgent)
	mockDB.EXPECT().Account().AnyTimes().Return(accountAgent)
	mockDB.EXPECT().Faucet().AnyTimes().Return(faucetAgent)
	searchFaucetsUC := mocks2.NewMockSearchFaucetsUseCase(ctrl)

	faucet := testdata.FakeFaucet()
	chain := testdata.FakeChain()
	account := testdata.FakeAccount()
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewRegisterFaucetUseCase(mockDB, searchFaucetsUC)
	faucet.TenantID = userInfo.TenantID

	expectedErr := errors.NotFoundError("error")
	t.Run("should execute use case successfully", func(t *testing.T) {
		searchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name},
			TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Faucet{}, nil)

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(chain, nil)
		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), faucet.CreditorAccount.String(), userInfo.AllowedTenants,
			userInfo.Username).Return(account, nil)

		faucetAgent.EXPECT().Insert(gomock.Any(), faucet).Return(faucet, nil)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, faucet, resp)
	})

	t.Run("should fail with AlreadyExistsError if search faucets returns results", func(t *testing.T) {
		searchFaucetsUC.EXPECT().
			Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name}, TenantID: userInfo.TenantID},
				userInfo).Return([]*entities.Faucet{faucet}, nil)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, errors.IsAlreadyExistsError(err))
	})

	t.Run("should fail with same error if search faucets fails", func(t *testing.T) {
		searchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name},
			TenantID: userInfo.TenantID}, userInfo).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})
	
	t.Run("should fail with invalid parameter error if chain is not found", func(t *testing.T) {
		searchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name},
			TenantID: userInfo.TenantID}, userInfo).Return([]*entities.Faucet{}, nil)

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(nil, expectedErr)
		
		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	
	t.Run("should fail with invalid parameter error if chain is not found", func(t *testing.T) {
		searchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name},
			TenantID: userInfo.TenantID}, userInfo).Return([]*entities.Faucet{}, nil)

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(nil, expectedErr)
		
		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	
	t.Run("should fail with invalid parameter error if account is not found", func(t *testing.T) {
		searchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name},
			TenantID: userInfo.TenantID}, userInfo).Return([]*entities.Faucet{}, nil)

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(chain, nil)
		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), faucet.CreditorAccount.String(), userInfo.AllowedTenants,
			userInfo.Username).Return(nil, expectedErr)
		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
	})

	t.Run("should fail with same error if insert faucet fails", func(t *testing.T) {
		searchFaucetsUC.EXPECT().Execute(gomock.Any(), &entities.FaucetFilters{Names: []string{faucet.Name}, TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Faucet{}, nil)
		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(chain, nil)
		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), faucet.CreditorAccount.String(), userInfo.AllowedTenants,
			userInfo.Username).Return(account, nil)
		faucetAgent.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(faucet, expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}
