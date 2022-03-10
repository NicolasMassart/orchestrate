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

func TestUpdateAccount_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	identityAgent := mocks.NewMockAccountAgent(ctrl)
	mockDB.EXPECT().Account().Return(identityAgent).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewUpdateAccountUseCase(mockDB)

	t.Run("should update identity successfully", func(t *testing.T) {
		idenEntity := testdata.FakeAccount()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address.Hex(), userInfo.AllowedTenants, userInfo.Username).Return(idenEntity, nil)

		idenEntity.Attributes = idenEntity.Attributes
		idenEntity.Alias = idenEntity.Alias
		identityAgent.EXPECT().Update(gomock.Any(), idenEntity).Return(idenEntity, nil)
		resp, err := usecase.Execute(ctx, idenEntity, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, resp.Attributes, idenEntity.Attributes)
		assert.Equal(t, resp.Alias, idenEntity.Alias)
	})

	t.Run("should update non empty identity values", func(t *testing.T) {
		idenEntity := testdata.FakeAccount()
		idenEntity.Attributes = nil
		idenEntity.Alias = ""

		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address.Hex(), userInfo.AllowedTenants, userInfo.Username).Return(idenEntity, nil)

		identityAgent.EXPECT().Update(gomock.Any(), idenEntity).Return(idenEntity, nil)
		resp, err := usecase.Execute(ctx, idenEntity, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, resp.Attributes, idenEntity.Attributes)
		assert.Equal(t, resp.Alias, idenEntity.Alias)
	})

	t.Run("should fail with same error if get identity fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		idenEntity := testdata.FakeAccount()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address.Hex(), userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, idenEntity, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateAccountComponent), err)
	})

	t.Run("should fail with same error if get identity fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		idenEntity := testdata.FakeAccount()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address.Hex(), userInfo.AllowedTenants, userInfo.Username).Return(idenEntity, nil)

		identityAgent.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
		_, err := usecase.Execute(ctx, idenEntity, userInfo)

		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateAccountComponent), err)
	})
}
