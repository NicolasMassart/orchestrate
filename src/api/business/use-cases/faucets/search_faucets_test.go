// +build unit

package faucets

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

func TestSearchFaucets_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	faucetAgent := mocks.NewMockFaucetAgent(ctrl)
	mockDB.EXPECT().Faucet().Return(faucetAgent).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewSearchFaucets(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		filters := &entities.FaucetFilters{
			Names:     []string{"name1", "name2"},
			ChainRule: "chainRule",
		}
		faucet := testdata.FakeFaucet()
		faucetAgent.EXPECT().Search(gomock.Any(), filters, userInfo.AllowedTenants).Return([]*entities.Faucet{faucet}, nil)

		resp, err := usecase.Execute(ctx, filters, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, []*entities.Faucet{faucet}, resp)
	})

	t.Run("should fail with same error if search faucets fails", func(t *testing.T) {
		expectedErr := errors.PostgresConnectionError("error")

		faucetAgent.EXPECT().Search(gomock.Any(), nil, userInfo.AllowedTenants).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, nil, userInfo)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchFaucetsComponent), err)
	})
}
