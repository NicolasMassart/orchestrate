package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/event_streams"
	"github.com/consensys/orchestrate/src/api/store"
)

type eventStreamUseCases struct {
	createUC usecases.CreateEventStreamUseCase
	searchUC usecases.SearchEventStreamsUseCase
}

func newEventStreamUseCases(db store.EventStreamAgent) *eventStreamUseCases {
	return &eventStreamUseCases{
		createUC: streams.NewCreateUseCase(db),
		searchUC: streams.NewSearchUseCase(db),
	}
}

func (u *accountUseCases) Search() usecases.SearchAccountsUseCase {
	return u.searchAccountsUC
}

func (u *accountUseCases) Create() usecases.CreateAccountUseCase {
	return u.createAccountUC
}
