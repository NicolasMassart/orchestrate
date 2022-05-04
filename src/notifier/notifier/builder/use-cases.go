package builder

import (
	"net/http"

	"github.com/consensys/orchestrate/src/infra/kafka"
	"github.com/consensys/orchestrate/src/notifier/notifier/use-cases"
	"github.com/consensys/orchestrate/src/notifier/notifier/use-cases/notifications"
	"github.com/consensys/orchestrate/src/notifier/store"
)

type useCases struct {
	create usecases.CreateTxNotificationUseCase
	send   usecases.SendNotificationUseCase
}

var _ usecases.UseCases = &useCases{}

func NewUseCases(db store.NotificationAgent, kafkaProducer kafka.Producer, webhookClient *http.Client) usecases.UseCases {
	return &useCases{
		create: notifications.NewCreateTransactionUseCase(db),
		send:   notifications.NewSendUseCase(db, kafkaProducer, webhookClient),
	}
}

func (u useCases) Send() usecases.SendNotificationUseCase {
	return u.send
}

func (u useCases) CreateTransaction() usecases.CreateTxNotificationUseCase {
	return u.create
}
