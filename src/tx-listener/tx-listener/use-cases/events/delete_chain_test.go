// +build unit

package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/tx-listener/tx-listener/use-cases/mocks"
	storemocks "github.com/consensys/orchestrate/src/tx-listener/store/mocks"

	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeleteChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chainStore := storemocks.NewMockChain(ctrl)
	retryJobSessionManager := mocks.NewMockRetryJobSessionManager(ctrl)
	pendingJobStore := storemocks.NewMockPendingJob(ctrl)
	logger := log.NewLogger()

	usecase := DeleteChainUseCase(retryJobSessionManager, chainStore, pendingJobStore, logger)

	t.Run("should delete chain successfully", func(t *testing.T) {
		chain := testdata.FakeChain()

		retryJobSessionManager.EXPECT().StopChainSessions(gomock.Any(), chain.UUID).Return(nil)
		pendingJobStore.EXPECT().DeletePerChainUUID(gomock.Any(), chain.UUID).Return(nil)
		chainStore.EXPECT().Delete(gomock.Any(), chain.UUID).Return(nil)

		err := usecase.Execute(ctx, chain.UUID)

		assert.NoError(t, err)
	})

	t.Run("should fail if delete chain returns an error", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := fmt.Errorf("expected_err")

		retryJobSessionManager.EXPECT().StopChainSessions(gomock.Any(), chain.UUID).Return(nil)
		pendingJobStore.EXPECT().DeletePerChainUUID(gomock.Any(), chain.UUID).Return(nil)
		chainStore.EXPECT().Delete(gomock.Any(), chain.UUID).Return(expectedErr)

		err := usecase.Execute(ctx, chain.UUID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	t.Run("should fail if delete pending jobs returns an error", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := fmt.Errorf("expected_err")

		retryJobSessionManager.EXPECT().StopChainSessions(gomock.Any(), chain.UUID).Return(nil)
		pendingJobStore.EXPECT().DeletePerChainUUID(gomock.Any(), chain.UUID).Return(expectedErr)

		err := usecase.Execute(ctx, chain.UUID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
	
	t.Run("should fail if stop chain sessions returns an error", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := fmt.Errorf("expected_err")

		retryJobSessionManager.EXPECT().StopChainSessions(gomock.Any(), chain.UUID).Return(expectedErr)

		err := usecase.Execute(ctx, chain.UUID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
