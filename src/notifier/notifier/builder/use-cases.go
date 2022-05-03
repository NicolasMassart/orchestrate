package builder

import (
	"github.com/consensys/orchestrate/src/infra/messenger"
	"github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
	"github.com/consensys/orchestrate/src/notifier/notifier/use-cases/notifications"
	"github.com/consensys/orchestrate/src/notifier/store"
)

type useCases struct {
	create usecases.CreateTxNotificationUseCase
	send   usecases.SendNotificationUseCase
}

var _ usecases.UseCases = &useCases{}

func NewUseCases(db store.NotificationAgent, kafkaNotifier, webhookNotifier messenger.Producer) usecases.UseCases {
	return &useCases{
		create: notifications.NewCreateTransactionUseCase(db),
		send:   notifications.NewSendUseCase(db, kafkaNotifier, webhookNotifier),
	}
}

func (u useCases) Send() usecases.SendNotificationUseCase {
	return u.send
}

func (u useCases) CreateTransaction() usecases.CreateTxNotificationUseCase {
	return u.create
}
