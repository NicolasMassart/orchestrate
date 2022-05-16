package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/faucets"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	qkmclient "github.com/consensys/quorum-key-manager/pkg/client"
)

type useCases struct {
	jobUseCases          usecases.JobUseCases
	scheduleUseCases     usecases.ScheduleUseCases
	transactionUseCases  usecases.TransactionUseCases
	faucetUseCases       usecases.FaucetUseCases
	chainUseCases        usecases.ChainUseCases
	contractUseCases     usecases.ContractUseCases
	accountUseCases      usecases.AccountUseCases
	eventStreamUseCases  usecases.EventStreamsUseCases
	subscriptionUseCases usecases.SubscriptionUseCases
	notificationUseCases usecases.NotificationsUseCases
}

func NewUseCases(
	db store.DB,
	appMetrics metrics.TransactionSchedulerMetrics,
	keyManagerClient qkmclient.EthClient,
	qkmStoreID string,
	ec ethclient.Client,
	messengerClient sdk.OrchestrateMessenger,
) usecases.UseCases {
	chainUseCases := newChainUseCases(db, ec)
	contractUseCases := newContractUseCases(db)
	faucetUseCases := newFaucetUseCases(db)
	getFaucetCandidateUC := faucets.NewGetFaucetCandidateUseCase(faucetUseCases.Search(), ec)
	scheduleUseCases := newScheduleUseCases(db)
	eventStreamUseCases := newEventStreamUseCases(db, contractUseCases, chainUseCases, messengerClient)
	subscriptionsUseCases := NewSubscriptionUseCases(db, contractUseCases, chainUseCases, eventStreamUseCases.Search(),
		messengerClient)
	jobUseCases := newJobUseCases(db, appMetrics, messengerClient, eventStreamUseCases, chainUseCases, qkmStoreID)
	transactionUseCases := newTransactionUseCases(db, chainUseCases.Search(), getFaucetCandidateUC,
		scheduleUseCases, jobUseCases, contractUseCases.Get())
	accountUseCases := newAccountUseCases(db, keyManagerClient, chainUseCases.Search(),
		transactionUseCases.Send(), getFaucetCandidateUC)

	return &useCases{
		jobUseCases:          jobUseCases,
		scheduleUseCases:     scheduleUseCases,
		transactionUseCases:  transactionUseCases,
		faucetUseCases:       faucetUseCases,
		chainUseCases:        chainUseCases,
		contractUseCases:     contractUseCases,
		accountUseCases:      accountUseCases,
		eventStreamUseCases:  eventStreamUseCases,
		subscriptionUseCases: subscriptionsUseCases,
		notificationUseCases: NewNotificationUseCases(db.Notification()),
	}
}

func (ucs *useCases) Jobs() usecases.JobUseCases {
	return ucs.jobUseCases
}

func (ucs *useCases) Schedules() usecases.ScheduleUseCases {
	return ucs.scheduleUseCases
}

func (ucs *useCases) Transactions() usecases.TransactionUseCases {
	return ucs.transactionUseCases
}

func (ucs *useCases) Faucets() usecases.FaucetUseCases {
	return ucs.faucetUseCases
}

func (ucs *useCases) Chains() usecases.ChainUseCases {
	return ucs.chainUseCases
}

func (ucs *useCases) Contracts() usecases.ContractUseCases {
	return ucs.contractUseCases
}

func (ucs *useCases) Accounts() usecases.AccountUseCases {
	return ucs.accountUseCases
}

func (ucs *useCases) EventStreams() usecases.EventStreamsUseCases {
	return ucs.eventStreamUseCases
}

func (ucs *useCases) Subscriptions() usecases.SubscriptionUseCases {
	return ucs.subscriptionUseCases
}

func (ucs *useCases) Notifications() usecases.NotificationsUseCases {
	return ucs.notificationUseCases
}
