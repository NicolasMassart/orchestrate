package chains

import (
	"context"
	"testing"

	testutils2 "github.com/ConsenSys/orchestrate/pkg/types/testutils"
	"github.com/ConsenSys/orchestrate/services/api/business/parsers"
	mocks2 "github.com/ConsenSys/orchestrate/services/api/business/use-cases/mocks"

	"github.com/ConsenSys/orchestrate/pkg/multitenancy"

	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/store/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockDBTX := mocks.NewMockTx(ctrl)
	chainAgent := mocks.NewMockChainAgent(ctrl)
	privateTxManagerAgent := mocks.NewMockPrivateTxManagerAgent(ctrl)
	getChainUC := mocks2.NewMockGetChainUseCase(ctrl)

	mockDB.EXPECT().Begin().Return(mockDBTX, nil).AnyTimes()
	mockDB.EXPECT().Chain().Return(chainAgent).AnyTimes()
	mockDBTX.EXPECT().Chain().Return(chainAgent).AnyTimes()
	mockDBTX.EXPECT().PrivateTxManager().Return(privateTxManagerAgent).AnyTimes()
	mockDBTX.EXPECT().Commit().Return(nil).AnyTimes()
	mockDBTX.EXPECT().Rollback().Return(nil).AnyTimes()
	mockDBTX.EXPECT().Close().Return(nil).AnyTimes()

	usecase := NewDeleteChainUseCase(mockDB, getChainUC)

	tenantID := multitenancy.DefaultTenant
	tenants := []string{tenantID}

	t.Run("should execute use case successfully", func(t *testing.T) {
		chain := testutils2.FakeChain()
		chainModel := parsers.NewChainModelFromEntity(chain)

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", tenants).Return(chain, nil)
		privateTxManagerAgent.EXPECT().Delete(gomock.Any(), chainModel.PrivateTxManagers[0]).Return(nil)
		chainAgent.EXPECT().Delete(gomock.Any(), chainModel, tenants).Return(nil)

		err := usecase.Execute(ctx, "uuid", tenants)

		assert.NoError(t, err)
	})

	t.Run("should fail with same error if get chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", tenants).Return(nil, expectedErr)

		err := usecase.Execute(ctx, "uuid", tenants)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteChainComponent), err)
	})

	t.Run("should fail with same error if delete private tx manager fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		chain := testutils2.FakeChain()
		chainModel := parsers.NewChainModelFromEntity(chain)

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", tenants).Return(chain, nil)
		privateTxManagerAgent.EXPECT().Delete(gomock.Any(), chainModel.PrivateTxManagers[0]).Return(expectedErr)

		err := usecase.Execute(ctx, "uuid", tenants)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteChainComponent), err)
	})

	t.Run("should fail with same error if delete chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		chain := testutils2.FakeChain()

		getChainUC.EXPECT().Execute(gomock.Any(), "uuid", tenants).Return(chain, nil)
		privateTxManagerAgent.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
		chainAgent.EXPECT().Delete(gomock.Any(), parsers.NewChainModelFromEntity(chain), tenants).Return(expectedErr)

		err := usecase.Execute(ctx, "uuid", tenants)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(deleteChainComponent), err)
	})
}
