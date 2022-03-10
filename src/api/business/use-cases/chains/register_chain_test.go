// +build unit

package chains

import (
	"context"
	"math/big"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/infra/ethclient/mock"
	"github.com/ethereum/go-ethereum/core/types"

	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)
	mockSearchChainsUC := mocks2.NewMockSearchChainsUseCase(ctrl)
	mockEthClient := mock.NewMockClient(ctrl)

	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()
	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")

	usecase := NewRegisterChainUseCase(mockDB, mockSearchChainsUC, mockEthClient)

	t.Run("should execute use case successfully", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Chain{}, nil)
		mockEthClient.EXPECT().Network(gomock.Any(), chain.URLs[0]).Return(big.NewInt(888), nil)
		chainAgent.EXPECT().Insert(gomock.Any(), chain).Return(nil)

		resp, err := usecase.Execute(ctx, chain, false, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, chain, resp)
	})

	t.Run("should execute use case successfully from latest block", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username
		chainTip := big.NewInt(1)

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID}, userInfo).
			Return([]*entities.Chain{}, nil)
		mockEthClient.EXPECT().Network(gomock.Any(), chain.URLs[0]).
			Return(big.NewInt(888), nil)
		mockEthClient.EXPECT().HeaderByNumber(gomock.Any(), chain.URLs[0], nil).
			Return(&types.Header{
				Number: chainTip,
			}, nil)
		chainAgent.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := usecase.Execute(ctx, chain, true, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, uint64(1), resp.ListenerStartingBlock)
		assert.Equal(t, uint64(1), resp.ListenerCurrentBlock)
	})

	t.Run("should execute use case successfully with private tx manager", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Chain{}, nil)
		mockEthClient.EXPECT().Network(gomock.Any(), chain.URLs[0]).Return(big.NewInt(888), nil)
		chainAgent.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := usecase.Execute(ctx, chain, false, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, chain, resp)
	})

	t.Run("should fail with AlreadyExistsError if search chains returns results", func(t *testing.T) {
		chain := testdata.FakeChain()

		mockSearchChainsUC.EXPECT().
			Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID}, userInfo).
			Return([]*entities.Chain{chain}, nil)

		resp, err := usecase.Execute(ctx, chain, false, userInfo)

		assert.Nil(t, resp)
		assert.True(t, errors.IsAlreadyExistsError(err))
	})

	t.Run("should fail with same error if search chains fails", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := errors.NotFoundError("error")
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID},
			userInfo).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, chain, false, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(registerChainComponent), err)
	})

	t.Run("should fail with same error if insert chain fails", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := errors.NotFoundError("error")
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username

		mockSearchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chain.Name}, TenantID: userInfo.TenantID},
			userInfo).Return([]*entities.Chain{}, nil)
		mockEthClient.EXPECT().Network(gomock.Any(), chain.URLs[0]).Return(big.NewInt(888), nil)
		chainAgent.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(expectedErr)

		resp, err := usecase.Execute(ctx, chain, false, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(registerChainComponent), err)
	})
}
