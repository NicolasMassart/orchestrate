// +build unit

package faucets

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/stretchr/testify/require"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFaucet_Execute(t *testing.T) {
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

	faucet := testdata.FakeFaucet()
	chain := testdata.FakeChain()
	account := testdata.FakeAccount()
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateFaucetUseCase(mockDB)
	faucet.TenantID = userInfo.TenantID

	expectedErr := errors.NotFoundError("error")
	t.Run("should execute use case successfully", func(t *testing.T) {
		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(chain, nil)
		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), faucet.CreditorAccount.String(), userInfo.AllowedTenants,
			userInfo.Username).Return(account, nil)
		faucetAgent.EXPECT().Update(gomock.Any(), faucet, userInfo.AllowedTenants).Return(faucet, nil)
		faucetAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.UUID, userInfo.AllowedTenants).Return(faucet, nil)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, faucet, resp)
	})
	
	t.Run("should fail with the same error if update faucet fails", func(t *testing.T) {
		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(chain, nil)
		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), faucet.CreditorAccount.String(), userInfo.AllowedTenants,
			userInfo.Username).Return(account, nil)
		faucetAgent.EXPECT().Update(gomock.Any(), faucet, userInfo.AllowedTenants).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	t.Run("should fail with invalid params error if cannot find account", func(t *testing.T) {
		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(chain, nil)
		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), faucet.CreditorAccount.String(), userInfo.AllowedTenants,
			userInfo.Username).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
	
	t.Run("should fail with invalid params error if cannot find chain", func(t *testing.T) {
		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), faucet.ChainRule, userInfo.AllowedTenants, userInfo.Username).
			Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, faucet, userInfo)

		assert.Nil(t, resp)
		require.Error(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
	})
}
