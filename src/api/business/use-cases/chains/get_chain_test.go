// +build unit

package chains

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

func TestGetChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)

	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()

	usecase := NewGetChainUseCase(mockDB)
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")

	t.Run("should execute use case successfully", func(t *testing.T) {
		chain := testdata.FakeChain()

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), chain.UUID, userInfo.AllowedTenants, userInfo.Username).Return(chain, nil)

		resp, err := usecase.Execute(ctx, chain.UUID, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, chain, resp)
	})

	t.Run("should fail with same error if get chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), "uuid", userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, "uuid", userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getChainComponent), err)
	})
}
