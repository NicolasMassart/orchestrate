// +build unit

package contracts

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/api/store/mocks"
)

func TestGetTags_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	contractName := "myContract"
	tagAgent := mocks.NewMockContractAgent(ctrl)
	usecase := NewGetTagsUseCase(tagAgent)

	t.Run("should execute use case successfully", func(t *testing.T) {
		tags := []string{"latest", "v1.0.0"}
		tagAgent.EXPECT().ListTags(gomock.Any(), contractName).Return(tags, nil)

		response, err := usecase.Execute(ctx, contractName)

		assert.Equal(t, response, tags)
		assert.NoError(t, err)
	})

	t.Run("should fail if data agent fails", func(t *testing.T) {
		dataAgentError := fmt.Errorf("error")
		tagAgent.EXPECT().ListTags(gomock.Any(), contractName).Return(nil, dataAgentError)

		response, err := usecase.Execute(ctx, contractName)

		assert.Nil(t, response)
		assert.Equal(t, errors.FromError(dataAgentError).ExtendComponent(getTagsComponent), err)
	})
}
