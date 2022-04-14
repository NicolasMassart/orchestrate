package builder

import (
	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/event_streams"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/infra/push_notification"
)

type eventStreamUseCases struct {
	create   usecases.CreateEventStreamUseCase
	search   usecases.SearchEventStreamsUseCase
	notifyTx usecases.NotifyTransactionUseCase
	get      usecases.GetEventStreamUseCase
	update   usecases.UpdateEventStreamUseCase
	delete   usecases.DeleteEventStreamUseCase
}

var _ usecases.EventStreamsUseCases = &eventStreamUseCases{}

func newEventStreamUseCases(db store.EventStreamAgent, notifier pushnotification.Notifier, contracts usecases.ContractUseCases, chains usecases.ChainUseCases) *eventStreamUseCases {
	return &eventStreamUseCases{
		get:      streams.NewGetUseCase(db),
		create:   streams.NewCreateUseCase(db, chains.Search()),
		search:   streams.NewSearchUseCase(db),
		notifyTx: streams.NewNotifyTransactionUseCase(db, notifier, contracts.Search(), contracts.DecodeLog()),
		update:   streams.NewUpdateUseCase(db, chains.Search()),
		delete:   streams.NewDeleteUseCase(db),
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

func (u *eventStreamUseCases) Get() usecases.GetEventStreamUseCase {
	return u.get
}

func (u *eventStreamUseCases) Update() usecases.UpdateEventStreamUseCase {
	return u.update
}

func (u *eventStreamUseCases) Delete() usecases.DeleteEventStreamUseCase {
	return u.delete
}
