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
	notifySub usecases.NotifySubscriptionEventUseCase
	get       usecases.GetSubscriptionUseCase
	update    usecases.UpdateSubscriptionUseCase
	delete    usecases.DeleteSubscriptionUseCase
}

var _ usecases.SubscriptionUseCases = &subscriptionUseCases{}

func NewSubscriptionUseCases(
	db store.SubscriptionAgent,
	contracts usecases.ContractUseCases,
	chains usecases.ChainUseCases,
	eventstreams usecases.EventStreamsUseCases,
	txListenerMessenger sdk.MessengerTxListener,
	notifierMessenger sdk.MessengerNotifier,
) *subscriptionUseCases {
	searchUC := subscriptions.NewSearchUseCase(db)
	return &subscriptionUseCases{
		get:       subscriptions.NewGetUseCase(db),
		create:    subscriptions.NewCreateUseCase(db, chains.Search(), contracts.Get(), eventstreams.Search(), txListenerMessenger),
		search:    searchUC,
		notifySub: subscriptions.NewNotifySubscriptionEventUseCase(db, searchUC, contracts.Search(), eventstreams.Get(), contracts.DecodeLog(), notifierMessenger),
		update:    subscriptions.NewUpdateUseCase(db, eventstreams.Search(), txListenerMessenger),
		delete:    subscriptions.NewDeleteUseCase(db, txListenerMessenger),
	}
}

func (u *subscriptionUseCases) Search() usecases.SearchSubscriptionUseCase {
	return u.search
}

func (u *subscriptionUseCases) Create() usecases.CreateSubscriptionUseCase {
	return u.create
}

func (u *subscriptionUseCases) NotifySubscription() usecases.NotifySubscriptionEventUseCase {
	return u.notifySub
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
