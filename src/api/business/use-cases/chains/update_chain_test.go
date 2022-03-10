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
	chainAgent := mocks.NewMockChainAgent(ctrl)
	mockGetChainUC := mocks2.NewMockGetChainUseCase(ctrl)

	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()
	expectedErr := errors.NotFoundError("error")
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")

	usecase := NewUpdateChainUseCase(mockDB, mockGetChainUC)

	t.Run("should execute use case successfully", func(t *testing.T) {
		chain := testdata.FakeChain()

		chainAgent.EXPECT().Update(gomock.Any(), chain, userInfo.AllowedTenants, userInfo.Username).Return(nil)
		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(chain, nil)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, chain, resp)
	})

	t.Run("should fail with same error if get chain fails", func(t *testing.T) {
		chain := testdata.FakeChain()

		chainAgent.EXPECT().Update(gomock.Any(), chain, userInfo.AllowedTenants, userInfo.Username).Return(nil)
		mockGetChainUC.EXPECT().Execute(gomock.Any(), chain.UUID, userInfo).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChainComponent), err)
	})

	t.Run("should fail with same error if update chain fails", func(t *testing.T) {
		chain := testdata.FakeChain()

		chainAgent.EXPECT().Update(gomock.Any(), chain, userInfo.AllowedTenants, userInfo.Username).Return(expectedErr)

		resp, err := usecase.Execute(ctx, chain, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateChainComponent), err)
	})
}
