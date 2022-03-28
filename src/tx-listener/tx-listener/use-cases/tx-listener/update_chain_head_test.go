package txlistener

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateChainHead_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	logger := log.NewLogger()

	expectedErr := fmt.Errorf("expected_err")

	usecase := UpdateChainHeadUseCase(apiClient, logger)

	t.Run("should update chain head successfully", func(t *testing.T) {
		chain := testdata.FakeChain()

		apiClient.EXPECT().UpdateChain(gomock.Any(), chain.UUID, &types.UpdateChainRequest{
			Listener: &types.UpdateListenerRequest{
				CurrentBlock: chain.ListenerCurrentBlock,
			}}).Return(&types.ChainResponse{}, nil)

		err := usecase.Execute(ctx, chain.UUID, chain.ListenerCurrentBlock)
		assert.NoError(t, err)
	})

	t.Run("should fail if get code hash fails", func(t *testing.T) {
		chain := testdata.FakeChain()

		apiClient.EXPECT().UpdateChain(gomock.Any(), chain.UUID, gomock.Any()).Return(&types.ChainResponse{}, expectedErr)

		err := usecase.Execute(ctx, chain.UUID, chain.ListenerCurrentBlock)
		assert.Error(t, err)
		assert.True(t, errors.IsDependencyFailureError(err))
	})
}
