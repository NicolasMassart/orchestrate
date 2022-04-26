// +build unit

package streams

import (
	"context"
	"testing"

	mocks2 "github.com/consensys/orchestrate/src/api/business/use-cases/mocks"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/mocks"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockEventStreamAgent(ctrl)
	searchChainsUC := mocks2.NewMockSearchChainsUseCase(ctrl)

	userInfo := multitenancy.NewUserInfo("tenantOne", "username")
	chainName := "chain"
	chain := testdata.FakeChain()

	usecase := NewCreateUseCase(mockDB, searchChainsUC)

	t.Run("should create new event stream successfully: Webhook", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()
		eventStream.TenantID = userInfo.TenantID
		eventStream.OwnerID = userInfo.Username

		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{}, nil)
		searchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return([]*entities.Chain{chain}, nil)
		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{TenantID: userInfo.TenantID, ChainUUID: chain.UUID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{}, nil)
		mockDB.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(eventStream, nil)

		resp, err := usecase.Execute(ctx, eventStream, chainName, userInfo)

		assert.NoError(t, err)
		assert.Equal(t, resp, eventStream)
	})

	t.Run("should fail with InvalidParameter if chain is not found", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{}, nil)
		searchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return([]*entities.Chain{}, nil)

		_, err := usecase.Execute(ctx, eventStream, chainName, userInfo)
		assert.True(t, errors.IsInvalidParameterError(err))
	})

	t.Run("should fail with same error if search chain fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{}, nil)
		searchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, eventStream, chainName, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createEventStreamComponent), err)
	})

	t.Run("should fail with same error if search event streams fails", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		eventStream := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, eventStream, chainName, userInfo)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createEventStreamComponent), err)
	})

	t.Run("should fail with AlreadyExistsError if search event streams with Name+TenantID returns values", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()
		foundEventStreamEntity := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{foundEventStreamEntity}, nil)

		_, err := usecase.Execute(ctx, eventStream, chainName, userInfo)
		assert.Error(t, err)
		assert.True(t, errors.IsAlreadyExistsError(err))
	})
	
	t.Run("should fail with AlreadyExistsError if search event streams with TenantID+Chain returns values", func(t *testing.T) {
		eventStream := testdata.FakeWebhookEventStream()
		foundEventStreamEntity := testdata.FakeWebhookEventStream()

		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{Names: []string{eventStream.Name}, TenantID: userInfo.TenantID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{}, nil)
		searchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return([]*entities.Chain{chain}, nil)
		mockDB.EXPECT().Search(
			gomock.Any(),
			&entities.EventStreamFilters{TenantID: userInfo.TenantID, ChainUUID: chain.UUID},
			userInfo.AllowedTenants,
			userInfo.Username,
		).Return([]*entities.EventStream{foundEventStreamEntity}, nil)

		_, err := usecase.Execute(ctx, eventStream, chainName, userInfo)
		assert.Error(t, err)
		assert.True(t, errors.IsAlreadyExistsError(err))
	})

	t.Run("should fail with same error if cannot insert event stream", func(t *testing.T) {
		expectedErr := errors.NotFoundError("error")
		eventStream := testdata.FakeWebhookEventStream()

		searchChainsUC.EXPECT().Execute(gomock.Any(), &entities.ChainFilters{Names: []string{chainName}}, userInfo).Return([]*entities.Chain{chain}, nil)
		mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return([]*entities.EventStream{}, nil)
		mockDB.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		_, err := usecase.Execute(ctx, eventStream, chainName, userInfo)

		assert.Error(t, err)
		assert.Equal(t, errors.FromError(expectedErr).ExtendComponent(createEventStreamComponent), err)
	})
}
