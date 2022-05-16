package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/event_streams"
	"github.com/consensys/orchestrate/src/api/store"
)

type eventStreamUseCases struct {
	create               usecases.CreateEventStreamUseCase
	search               usecases.SearchEventStreamsUseCase
	notifyTx             usecases.NotifyTransactionUseCase
	notifyContractEvents usecases.NotifyContractEventsUseCase
	get                  usecases.GetEventStreamUseCase
	update               usecases.UpdateEventStreamUseCase
	delete               usecases.DeleteEventStreamUseCase
}

var _ usecases.EventStreamsUseCases = &eventStreamUseCases{}

func newEventStreamUseCases(
	db store.DB,
	contracts usecases.ContractUseCases,
	chains usecases.ChainUseCases,
	txNotifierMessenger sdk.MessengerNotifier,
) *eventStreamUseCases {
	return &eventStreamUseCases{
		get:                  streams.NewGetUseCase(db.EventStream()),
		create:               streams.NewCreateUseCase(db.EventStream(), chains.Search()),
		search:               streams.NewSearchUseCase(db.EventStream()),
		notifyTx:             streams.NewNotifyTransactionUseCase(db, contracts.Search(), contracts.DecodeLog(), txNotifierMessenger),
		notifyContractEvents: streams.NewNotifyContractEventsUseCase(db, contracts.Search(), contracts.DecodeLog(), txNotifierMessenger),
		update:               streams.NewUpdateUseCase(db.EventStream()),
		delete:               streams.NewDeleteUseCase(db.EventStream()),
	}
}

func (u *eventStreamUseCases) Search() usecases.SearchEventStreamsUseCase {
	return u.search
}

func (u *eventStreamUseCases) Create() usecases.CreateEventStreamUseCase {
	return u.create
}

func (u *eventStreamUseCases) NotifyTransaction() usecases.NotifyTransactionUseCase {
	return u.notifyTx
}

func (u *eventStreamUseCases) NotifyContractEvents() usecases.NotifyContractEventsUseCase {
	return u.notifyContractEvents
}

func (u *eventStreamUseCases) Get() usecases.GetEventStreamUseCase {
	return u.get
}

func (u *eventStreamUseCases) Update() usecases.UpdateEventStreamUseCase {
	return u.update
}

func (u *eventStreamUseCases) Delete() usecases.DeleteEventStreamUseCase {
	return u.delete
}
