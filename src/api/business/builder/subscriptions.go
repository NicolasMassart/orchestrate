package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/subscriptions"
	"github.com/consensys/orchestrate/src/api/store"
)

type subscriptionUseCases struct {
	create    usecases.CreateSubscriptionUseCase
	search    usecases.SearchSubscriptionUseCase
	get       usecases.GetSubscriptionUseCase
	update    usecases.UpdateSubscriptionUseCase
	delete    usecases.DeleteSubscriptionUseCase
}

var _ usecases.SubscriptionUseCases = &subscriptionUseCases{}

func NewSubscriptionUseCases(
	db store.DB,
	contracts usecases.ContractUseCases,
	chains usecases.ChainUseCases,
	searchEventStreams usecases.SearchEventStreamsUseCase,
	messengerClient sdk.OrchestrateMessenger,
) *subscriptionUseCases {
	searchUC := subscriptions.NewSearchUseCase(db.Subscription())
	return &subscriptionUseCases{
		get:       subscriptions.NewGetUseCase(db.Subscription()),
		create:    subscriptions.NewCreateUseCase(db.Subscription(), chains.Search(), contracts.Get(), searchEventStreams, messengerClient),
		search:    searchUC,
		update:    subscriptions.NewUpdateUseCase(db.Subscription(), searchEventStreams, messengerClient),
		delete:    subscriptions.NewDeleteUseCase(db.Subscription(), messengerClient),
	}
}

func (u *subscriptionUseCases) Search() usecases.SearchSubscriptionUseCase {
	return u.search
}

func (u *subscriptionUseCases) Create() usecases.CreateSubscriptionUseCase {
	return u.create
}

func (u *subscriptionUseCases) Get() usecases.GetSubscriptionUseCase {
	return u.get
}

func (u *subscriptionUseCases) Update() usecases.UpdateSubscriptionUseCase {
	return u.update
}

func (u *subscriptionUseCases) Delete() usecases.DeleteSubscriptionUseCase {
	return u.delete
}
