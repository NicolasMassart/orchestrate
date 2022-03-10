// +build unit

package chains

import (
	"context"
	"testing"

	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)
	getChainUC := mocks2.NewMockGetChainUseCase(ctrl)

	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()

	usecase := NewDeleteChainUseCase(mockDB, getChainUC)
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")

	t.Run("should execute use case successfully", func(t *testing.T) {
		chain := testdata.FakeChain()
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", userInfo).Return(chain, nil)
		chainAgent.EXPECT().Delete(gomock.Any(), chain, userInfo.AllowedTenants).Return(nil)

		err := usecase.Execute(ctx, "uuid", userInfo)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if get chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", userInfo).Return(nil, expectedErr)

		err := usecase.Execute(ctx, "uuid", userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteChainComponent), err)
	})

	t.Run("should fail with same error if delete chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		chain := testdata.FakeChain()
		chain.TenantID = userInfo.TenantID
		chain.OwnerID = userInfo.Username

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", userInfo).Return(chain, nil)
		chainAgent.EXPECT().Delete(gomock.Any(), chain, userInfo.AllowedTenants).Return(expectedErr)

		err := usecase.Execute(ctx, "uuid", userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteChainComponent), err)
	})
}
