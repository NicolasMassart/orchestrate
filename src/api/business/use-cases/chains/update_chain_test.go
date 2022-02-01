// +build unit

package chains

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"

	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockDBTX := mocks.NewMockTx(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)
	mockGetChainUC := mocks2.NewMockGetChainUseCase(ctrl)
	privateTxManagerAgent := mocks.NewMockPrivateTxManagerAgent(ctrl)

	mockDB.EXPECT().Begin().Return(mockDBTX, nil).AnyTimes()
	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()
	mockDBTX.EXPECT().Chain().Return(chainAgent).AnyTimes()
	mockDBTX.EXPECT().PrivateTxManager().Return(privateTxManagerAgent).AnyTimes()
	mockDBTX.EXPECT().Commit().Return(nil).AnyTimes()
	mockDBTX.EXPECT().Rollback().Return(nil).AnyTimes()
	mockDBTX.EXPECT().Close().Return(nil).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")

	usecase := NewUpdateChainUseCase(mockDB, mockGetChainUC)

	t.Run("should execute use case successfully", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.PrivateTxManager = nil

		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(chain, nil)
		chainAgent.EXPECT().Update(gomock.Any(), chain, userInfo.AllowedTenants, userInfo.Username).Return(nil)
		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(chain, nil)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, chain, resp)
	})

	t.Run("should execute use case successfully with private tx manager", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.PrivateTxManager = nil
		newChain := testdata.FakeChain()
		newChain.UUID = chain.UUID
		newChain.PrivateTxManager.UUID = chain.UUID

		mockGetChainUC.EXPECT().Execute(gomock.Any(), newChain.UUID, userInfo).Return(chain, nil)
		chainAgent.EXPECT().Update(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username).Return(nil)
		privateTxManagerAgent.EXPECT().Insert(gomock.Any(), newChain.PrivateTxManager).Return(nil)
		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(chain, nil)

		resp, err := usecase.Execute(ctx, newChain, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, chain, resp)
	})

	t.Run("should execute use case successfully with private tx manager to insert", func(t *testing.T) {
		chainUpdate := testdata.FakeChain()
		chainRetrieved := testdata.FakeChain()
		chainRetrieved.PrivateTxManager = nil
		chainRetrieved.UUID = chainUpdate.UUID
		chainUpdate.PrivateTxManager.ChainUUID = chainRetrieved.UUID

		mockGetChainUC.EXPECT().Execute(gomock.Any(), chainUpdate.UUID, userInfo).Return(chainRetrieved, nil)
		privateTxManagerAgent.EXPECT().Insert(gomock.Any(), chainUpdate.PrivateTxManager).Return(nil)
		chainAgent.EXPECT().Update(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username).Return(nil)
		mockGetChainUC.EXPECT().Execute(gomock.Any(), chainUpdate.UUID, userInfo).Return(chainRetrieved, nil)

		_, err := usecase.Execute(ctx, chainUpdate, userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if get chain fails", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := errors.NotFoundError("error")

		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChainComponent), err)
	})

	t.Run("should fail with same error if update private tx manager fails", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := errors.NotFoundError("error")

		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(chain, nil)
		privateTxManagerAgent.EXPECT().Update(gomock.Any(), gomock.Any()).Return(expectedErr)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChainComponent), err)
	})

	t.Run("should fail with same error if update chain fails", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := errors.NotFoundError("error")

		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(chain, nil)
		privateTxManagerAgent.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		chainAgent.EXPECT().Update(gomock.Any(), gomock.Any(), userInfo.AllowedTenants, userInfo.Username).Return(expectedErr)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChainComponent), err)
	})
}
