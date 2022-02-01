// +build unit

package chains

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSearchChains_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)

	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()

	usecase := NewSearchChainsUseCase(mockDB)
	userInfo := multitenancy.NewUserInfo("tenantOne", "username")

	t.Run("should execute use case successfully", func(t *testing.T) {
		filters := &entities.ChainFilters{
			Names: []string{"name1", "name2"},
		}
		chain := testdata.FakeChain()

		chainAgent.EXPECT().Search(gomock.Any(), filters, userInfo.AllowedTenants, userInfo.Username).Return([]*entities.Chain{chain}, nil)

		resp, err := usecase.Execute(ctx, filters, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, []*entities.Chain{chain}, resp)
	})

	t.Run("should fail with same error if search chains fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")

		chainAgent.EXPECT().Search(gomock.Any(), nil, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, nil, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchChainsComponent), err)
	})
}
