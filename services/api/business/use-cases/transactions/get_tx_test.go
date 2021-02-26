// +build unit

package transactions

import (
	"context"
	"fmt"
	testutils2 "github.com/ConsenSys/orchestrate/pkg/types/testutils"
	mocks2 "github.com/ConsenSys/orchestrate/services/api/business/use-cases/mocks"
	"github.com/ConsenSys/orchestrate/services/api/store/models/testutils"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ConsenSys/orchestrate/pkg/errors"
	"github.com/ConsenSys/orchestrate/services/api/store/mocks"
)

func TestGetTx_Execute(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockTransactionRequestDA := mocks.NewMockTransactionRequestAgent(ctrl)
	mockGetScheduleUC := mocks2.NewMockGetScheduleUseCase(ctrl)
	tenants := []string{"tenantID"}

	mockDB.EXPECT().TransactionRequest().Return(mockTransactionRequestDA).AnyTimes()

	usecase := NewGetTxUseCase(mockDB, mockGetScheduleUC)

	t.Run("should execute use case successfully", func(t *testing.T) {
		txRequest := testutils.FakeTxRequest(0)
		schedule := testutils2.FakeSchedule()

		mockTransactionRequestDA.EXPECT().FindOneByUUID(gomock.Any(), txRequest.Schedule.UUID, tenants).Return(txRequest, nil)
		mockGetScheduleUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, tenants).Return(schedule, nil)

		result, err := usecase.Execute(ctx, txRequest.Schedule.UUID, tenants)

		assert.NoError(t, err)
		assert.Equal(t, txRequest.IdempotencyKey, result.IdempotencyKey)
		assert.Equal(t, txRequest.ChainName, result.ChainName)
		assert.Equal(t, txRequest.CreatedAt, result.CreatedAt)
		assert.Equal(t, txRequest.Params, result.Params)
		assert.Equal(t, schedule, result.Schedule)
	})

	t.Run("should fail with same error if FindOneByUUID fails", func(t *testing.T) {
		uuid := "uuid"
		expectedErr := errors.NotFoundError("error")

		mockTransactionRequestDA.EXPECT().FindOneByUUID(gomock.Any(), uuid, tenants).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, uuid, tenants)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getTxComponent), err)
	})

	t.Run("should fail with same error if GetScheduleUseCase fails", func(t *testing.T) {
		txRequest := testutils.FakeTxRequest(0)
		expectedErr := fmt.Errorf("error")

		mockTransactionRequestDA.EXPECT().FindOneByUUID(gomock.Any(), txRequest.Schedule.UUID, tenants).Return(txRequest, nil)
		mockGetScheduleUC.EXPECT().Execute(gomock.Any(), txRequest.Schedule.UUID, tenants).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, txRequest.Schedule.UUID, tenants)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(getTxComponent), err)
	})
}
