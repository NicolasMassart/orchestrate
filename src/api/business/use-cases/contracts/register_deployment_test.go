// +build unit

package contracts

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSetCodeHash_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	codeHash := utils.StringToHexBytes("0xAB")
	contractAgent := mocks.NewMockContractAgent(ctrl)
	usecase := NewRegisterDeploymentUseCase(contractAgent)

	t.Run("should execute use case successfully", func(t *testing.T) {
		contractAgent.EXPECT().RegisterDeployment(gomock.Any(), chainID, contractAddress, codeHash).Return(nil)

		err := usecase.Execute(ctx, chainID, contractAddress, codeHash)

		assert.NoError(t, err)
	})

	t.Run("should fail if data agent fails", func(t *testing.T) {
		dataAgentError := fmt.Errorf("error")
		contractAgent.EXPECT().RegisterDeployment(gomock.Any(), chainID, contractAddress, codeHash).Return(dataAgentError)

		err := usecase.Execute(ctx, chainID, contractAddress, codeHash)

		assert.Equal(t, errors.FromError(dataAgentError).ExtendComponent(registerDeploymentComponent), err)
	})
}
