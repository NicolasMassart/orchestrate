// +build unit

package transactions

import (
	"context"
	"fmt"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
)

func TestSearchTxs_Execute(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockTransactionRequestDA := mocks.NewMockTransactionRequestAgent(ctrl)
	mockGetTxUC := mocks2.NewMockGetTxUseCase(ctrl)
	filter := &entities.TransactionRequestFilters{}

	mockDB.EXPECT().TransactionRequest().Return(mockTransactionRequestDA).AnyTimes()

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	usecase := NewSearchTransactionsUseCase(mockDB, mockGetTxUC)

	t.Run("should execute use case successfully", func(t *testing.T) {
		txRequest0 := testdata.FakeTxRequest()
		txRequest1 := testdata.FakeTxRequest()
		txRequests := []*entities.TxRequest{txRequest0, txRequest1}

		mockTransactionRequestDA.EXPECT().Search(gomock.Any(), filter, userInfo.AllowedTenants, userInfo.Username).Return(txRequests, nil)
		mockGetTxUC.EXPECT().Execute(gomock.Any(), txRequests[0].Schedule.UUID, userInfo).Return(txRequest0, nil)
		mockGetTxUC.EXPECT().Execute(gomock.Any(), txRequests[1].Schedule.UUID, userInfo).Return(txRequest1, nil)

		result, err := usecase.Execute(ctx, filter, userInfo)

		assert.Nil(t, err)

		assert.Equal(t, txRequest0.Schedule.UUID, result[0].Schedule.UUID)
		assert.Equal(t, txRequest0.IdempotencyKey, result[0].IdempotencyKey)
		assert.Equal(t, txRequest0.CreatedAt, result[0].CreatedAt)
		assert.Equal(t, txRequest0.Params, result[0].Params)

		assert.Equal(t, txRequest1.Schedule.UUID, result[1].Schedule.UUID)
		assert.Equal(t, txRequest1.IdempotencyKey, result[1].IdempotencyKey)
		assert.Equal(t, txRequest1.CreatedAt, result[1].CreatedAt)
		assert.Equal(t, txRequest1.Params, result[1].Params)
	})

	t.Run("should fail with same error if Search fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")

		mockTransactionRequestDA.EXPECT().Search(gomock.Any(), filter, userInfo.AllowedTenants, userInfo.Username).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, filter, userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchTxsComponent), err)
	})

	t.Run("should fail with same error if GetTxUseCase fails", func(t *testing.T) {
		txRequests := []*entities.TxRequest{testdata.FakeTxRequest()}
		expectedErr := fmt.Errorf("error")

		mockTransactionRequestDA.EXPECT().Search(gomock.Any(), filter, userInfo.AllowedTenants, userInfo.Username).Return(txRequests, nil)
		mockGetTxUC.EXPECT().Execute(gomock.Any(), txRequests[0].Schedule.UUID, userInfo).Return(nil, expectedErr)

		response, err := usecase.Execute(ctx, filter, userInfo)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(searchTxsComponent), err)
	})
}
