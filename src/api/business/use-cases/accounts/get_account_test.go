// +build unit

package accounts

import (
	"context"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAccount_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	accountAgent := mocks.NewMockAccountAgent(ctrl)
	mockDB.EXPECT().Account().Return(accountAgent).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewGetAccountUseCase(mockDB)

	t.Run("should execute use case successfully", func(t *testing.T) {
		iden := testdata.FakeAccount()

		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), iden.Address.String(), userInfo.AllowedTenants, userInfo.Username).
			Return(iden, nil)

		resp, err := usecase.Execute(ctx, iden.Address, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, iden, resp)
	})

	t.Run("should fail with same error if get account fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		acc := testdata.FakeAccount()

		accountAgent.EXPECT().FindOneByAddress(gomock.Any(), acc.Address.String(), userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, acc.Address, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getAccountComponent), err)
	})
}
