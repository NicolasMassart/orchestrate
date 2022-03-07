// +build unit

package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	storemocks "github.com/consensys/orchestrate/src/chain-listener/store/mocks"

	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAddChain_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chainStore := storemocks.NewMockChain(ctrl)
	logger := log.NewLogger()

	usecase := AddChainUseCase(chainStore, logger)

	t.Run("should add new chain successfully", func(t *testing.T) {
		chain := testdata.FakeChain()

		chainStore.EXPECT().Add(gomock.Any(), chain).Return(nil)

		err := usecase.Execute(ctx, chain)

		assert.NoError(t, err)
	})

	t.Run("should fail if add chain returns an error", func(t *testing.T) {
		chain := testdata.FakeChain()
		expectedErr := fmt.Errorf("expected_err")

		chainStore.EXPECT().Add(gomock.Any(), chain).Return(expectedErr)

		err := usecase.Execute(ctx, chain)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
