package chains

import (
	"context"
	"testing"

	"github.com/ConsenSys/orchestrate/pkg/multitenancy"

	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/business/parsers"
	"github.com/ConsenSys/orchestrate/services/api/store/mocks"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
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

	tenantID := multitenancy.DefaultTenant
	tenants := []string{tenantID}

	t.Run("should execute use case successfully", func(t *testing.T) {
		chainModel := testutils.FakeChainModel()

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), chainModel.UUID, tenants).Return(chainModel, nil)

		resp, err := usecase.Execute(ctx, chainModel.UUID, tenants)

		assert.NoError(t, err)
		assert.Equal(t, parsers.NewChainFromModel(chainModel), resp)
	})

	t.Run("should fail with same error if get chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		chainAgent.EXPECT().FindOneByUUID(gomock.Any(), "uuid", tenants).Return(nil, expectedErr)

		resp, err := usecase.Execute(ctx, "uuid", tenants)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getChainComponent), err)
	})
}
