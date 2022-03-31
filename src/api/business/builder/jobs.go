package builder

import (
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/jobs"
	"github.com/consensys/orchestrate/src/api/business/use-cases/notifications"
	"github.com/consensys/orchestrate/src/api/metrics"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/infra/kafka"
	broker "github.com/consensys/orchestrate/src/infra/kafka/sarama"
)

type jobUseCases struct {
	createJob   usecases.CreateJobUseCase
	getJob      usecases.GetJobUseCase
	startJob    usecases.StartJobUseCase
	resendJobTx usecases.ResendJobTxUseCase
	retryJobTx  usecases.RetryJobTxUseCase
	updateJob   usecases.UpdateJobUseCase
	searchJobs  usecases.SearchJobsUseCase
}

func newJobUseCases(
	db store.DB,
	appMetrics metrics.TransactionSchedulerMetrics,
	producer kafka.Producer,
	topicsCfg *broker.TopicConfig,
	getChainUC usecases.GetChainUseCase,
	searchContractUC usecases.SearchContractUseCase,
	decodeLogUC usecases.DecodeEventLogUseCase,
	qkmStoreID string,
) *jobUseCases {
	notifyMinedJobUC := notifications.NewNotifyMinedJobUseCase(db, searchContractUC, decodeLogUC, producer, topicsCfg)
	notifyFailedJobUC := notifications.NewNotifyFailedJobUseCase(db, producer, topicsCfg)
	startJobUC := jobs.NewStartJobUseCase(db, producer, topicsCfg, appMetrics)
	startNextJobUC := jobs.NewStartNextJobUseCase(db, startJobUC)
	createJobUC := jobs.NewCreateJobUseCase(db, getChainUC, qkmStoreID)

	return &jobUseCases{
		createJob:   createJobUC,
		getJob:      jobs.NewGetJobUseCase(db),
		searchJobs:  jobs.NewSearchJobsUseCase(db),
		updateJob:   jobs.NewUpdateJobUseCase(db, startNextJobUC, notifyMinedJobUC, notifyFailedJobUC, appMetrics),
		startJob:    startJobUC,
		resendJobTx: jobs.NewResendJobTxUseCase(db, producer, topicsCfg),
		retryJobTx:  jobs.NewRetryJobTxUseCase(db, createJobUC, startJobUC),
	}
}

func (u *jobUseCases) CreateJob() usecases.CreateJobUseCase {
	return u.createJob
}

func (u *jobUseCases) GetJob() usecases.GetJobUseCase {
	return u.getJob
}

func (u *jobUseCases) SearchJobs() usecases.SearchJobsUseCase {
	return u.searchJobs
}

func (u *jobUseCases) StartJob() usecases.StartJobUseCase {
	return u.startJob
}

func (u *jobUseCases) ResendJobTx() usecases.ResendJobTxUseCase {
	return u.resendJobTx
}

func (u *jobUseCases) RetryJobTx() usecases.RetryJobTxUseCase {
	return u.retryJobTx
}

func (u *jobUseCases) UpdateJob() usecases.UpdateJobUseCase {
	return u.updateJob
}
