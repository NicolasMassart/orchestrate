package usecases

type UseCases interface {
	Jobs() JobUseCases
	Schedules() ScheduleUseCases
	Transactions() TransactionUseCases
	Faucets() FaucetUseCases
	Chains() ChainUseCases
	Contracts() ContractUseCases
	Accounts() AccountUseCases
	EventStreams() EventStreamsUseCases
	Subscriptions() SubscriptionUseCases
	Notifications() NotificationsUseCases
}
