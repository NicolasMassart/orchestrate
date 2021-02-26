// +build unit

package accounts

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ConsenSys/orchestrate/pkg/errors"
	testutils3 "github.com/ConsenSys/orchestrate/pkg/types/testutils"
	"github.com/ConsenSys/orchestrate/services/api/store/mocks"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
)

func TestUpdateAccount_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	identityAgent := mocks.NewMockAccountAgent(ctrl)
	mockDB.EXPECT().Account().Return(identityAgent).AnyTimes()

	usecase := NewUpdateAccountUseCase(mockDB)

	tenantID := "tenantID"
	tenants := []string{tenantID}

	t.Run("should update identity successfully", func(t *testing.T) {
		idenEntity := testutils3.FakeAccount()
		idenModel := testutils.FakeAccountModel()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address, tenants).Return(idenModel, nil)

		idenModel.Attributes = idenEntity.Attributes
		idenModel.Alias = idenEntity.Alias
		identityAgent.EXPECT().Update(gomock.Any(), idenModel).Return(nil)
		resp, err := usecase.Execute(ctx, idenEntity, tenants)

		assert.NoError(t, err)
		assert.Equal(t, resp.Attributes, idenEntity.Attributes)
		assert.Equal(t, resp.Alias, idenEntity.Alias)
	})

	t.Run("should update non empty identity values", func(t *testing.T) {
		idenEntity := testutils3.FakeAccount()
		idenEntity.Attributes = nil
		idenEntity.Alias = ""

		idenModel := testutils.FakeAccountModel()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address, tenants).Return(idenModel, nil)

		identityAgent.EXPECT().Update(gomock.Any(), idenModel).Return(nil)
		resp, err := usecase.Execute(ctx, idenEntity, tenants)

		assert.NoError(t, err)
		assert.Equal(t, resp.Attributes, idenModel.Attributes)
		assert.Equal(t, resp.Alias, idenModel.Alias)
	})

	t.Run("should fail with same error if get identity fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		idenEntity := testutils3.FakeAccount()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address, tenants).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, idenEntity, tenants)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateAccountComponent), err)
	})

	t.Run("should fail with same error if get identity fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		idenEntity := testutils3.FakeAccount()
		idenModel := testutils.FakeAccountModel()
		identityAgent.EXPECT().FindOneByAddress(gomock.Any(), idenEntity.Address, tenants).Return(idenModel, nil)

		identityAgent.EXPECT().Update(gomock.Any(), gomock.Any()).Return(expectedErr)
		_, err := usecase.Execute(ctx, idenEntity, tenants)

		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(updateAccountComponent), err)
	})
}
