// +build unit

package accounts

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/api/business/use-cases/mocks"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	faucetNotFoundErr = errors.NotFoundError("not found faucet candidate")
)

func TestFundingAccount_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearchChainsUC := mocks.NewMockSearchChainsUseCase(ctrl)
	mockGetFaucetCandidate := mocks.NewMockGetFaucetCandidateUseCase(ctrl)
	mockSendTxUC := mocks.NewMockSendTxUseCase(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewFundAccountUseCase(mockSearchChainsUC, mockSendTxUC, mockGetFaucetCandidate)

	t.Run("should trigger funding identity successfully", func(t *testing.T) {
		account := testdata.FakeAccount()
		chains := []*entities.Chain{testdata.FakeChain()}
		faucet := testdata.FakeFaucet()
		chainName := "besu"

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).
			Return(chains, nil)
		mockGetFaucetCandidate.EXPECT().Execute(gomock.Any(), account.Address, chains[0], userInfo).Return(faucet, nil)
		mockSendTxUC.EXPECT().Execute(gomock.Any(), gomock.Any(), nil, multitenancy.NewInternalAdminUser()).Return(nil, nil)

		err := usecase.Execute(ctx, account, chainName, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should do nothing if there is not faucet candidates", func(t *testing.T) {
		account := testdata.FakeAccount()
		chains := []*entities.Chain{testdata.FakeChain()}
		chainName := "besu"

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return(chains, nil)
		mockGetFaucetCandidate.EXPECT().Execute(gomock.Any(), account.Address, chains[0], userInfo).Return(nil, faucetNotFoundErr)

		err := usecase.Execute(ctx, account, chainName, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with InvalidParameter if no chains are found", func(t *testing.T) {
		account := testdata.FakeAccount()
		chainName := "besu"

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return([]*entities.Chain{}, nil)

		err := usecase.Execute(ctx, account, chainName, userInfo)

		assert.Error(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
	})

	t.Run("should fail with same error if search chain fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")
		account := testdata.FakeAccount()
		chainName := "besu"

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return(nil, expectedErr)

		err := usecase.Execute(ctx, account, chainName, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(fundAccountComponent), err)
	})

	t.Run("should fail with same error if get faucet candidate fails", func(t *testing.T) {
		expectedErr := errors.ConnectionError("error")
		account := testdata.FakeAccount()
		chains := []*entities.Chain{testdata.FakeChain()}
		chainName := "besu"

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return(chains, nil)
		mockGetFaucetCandidate.EXPECT().
			Execute(gomock.Any(), account.Address, gomock.Any(), userInfo).
			Return(nil, expectedErr)

		err := usecase.Execute(ctx, account, chainName, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(fundAccountComponent), err)
	})

	t.Run("should fail with same error if send funding transaction fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		account := testdata.FakeAccount()
		chains := []*entities.Chain{testdata.FakeChain()}
		faucet := testdata.FakeFaucet()
		chainName := "besu"

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return(chains, nil)
		mockGetFaucetCandidate.EXPECT().Execute(gomock.Any(), account.Address, gomock.Any(), userInfo).Return(faucet, nil)
		mockSendTxUC.EXPECT().Execute(gomock.Any(), gomock.Any(), nil, multitenancy.NewInternalAdminUser()).Return(nil, expectedErr)

		err := usecase.Execute(ctx, account, chainName, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(fundAccountComponent), err)
	})
}
