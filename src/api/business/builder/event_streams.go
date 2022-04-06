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
}

func newEventStreamUseCases(db store.EventStreamAgent, notifier pushnotification.Notifier, contracts usecases.ContractUseCases) *eventStreamUseCases {
	return &eventStreamUseCases{
		create:   streams.NewCreateUseCase(db),
		search:   streams.NewSearchUseCase(db),
		notifyTx: streams.NewNotifyTransactionUseCase(db, notifier, contracts.Search(), contracts.DecodeLog()),
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
