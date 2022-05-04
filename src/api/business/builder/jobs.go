package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/jobs"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/store"
)

type jobUseCases struct {
	create   usecases.CreateJobUseCase
	get      usecases.GetJobUseCase
	start    usecases.StartJobUseCase
	resendTx usecases.ResendJobTxUseCase
	retryTx  usecases.RetryJobTxUseCase
	update   usecases.UpdateJobUseCase
	search   usecases.SearchJobsUseCase
}

func newJobUseCases(
	db store.DB,
	appMetrics metrics.TransactionSchedulerMetrics,
	txSenderMessenger sdk.MessengerTxSender,
	txListenerMessenger sdk.MessengerTxListener,
	eventStreams usecases.EventStreamsUseCases,
	chains usecases.ChainUseCases,
	qkmStoreID string,
) *jobUseCases {
	startJobUC := jobs.NewStartJobUseCase(db, txSenderMessenger, appMetrics)
	startNextJobUC := jobs.NewStartNextJobUseCase(db, startJobUC)
	createJobUC := jobs.NewCreateJobUseCase(db, chains.Get(), qkmStoreID)

	return &jobUseCases{
		create:   createJobUC,
		get:      jobs.NewGetJobUseCase(db),
		search:   jobs.NewSearchJobsUseCase(db),
		update:   jobs.NewUpdateJobUseCase(db, startNextJobUC, appMetrics, eventStreams.NotifyTransaction(), txListenerMessenger),
		start:    startJobUC,
		resendTx: jobs.NewResendJobTxUseCase(db, txSenderMessenger),
		retryTx:  jobs.NewRetryJobTxUseCase(db, createJobUC, startJobUC),
	}
}

func (u *jobUseCases) Create() usecases.CreateJobUseCase {
	return u.create
}

func (u *jobUseCases) Get() usecases.GetJobUseCase {
	return u.get
}

func (u *jobUseCases) Search() usecases.SearchJobsUseCase {
	return u.search
}

func (u *jobUseCases) Start() usecases.StartJobUseCase {
	return u.start
}

func (u *jobUseCases) ResendTx() usecases.ResendJobTxUseCase {
	return u.resendTx
}

func (u *jobUseCases) RetryTx() usecases.RetryJobTxUseCase {
	return u.retryTx
}

func (u *jobUseCases) Update() usecases.UpdateJobUseCase {
	return u.update
}
