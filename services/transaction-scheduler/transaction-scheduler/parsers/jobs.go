package parsers

import (
	"math/big"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/tx"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/store/models"
)

func NewJobModelFromEntities(job *types.Job, scheduleID *int) *models.Job {
	jobModel := &models.Job{
		UUID:        job.UUID,
		ChainUUID:   job.ChainUUID,
		Type:        job.Type,
		Labels:      job.Labels,
		Annotations: job.Annotations,
		ScheduleID:  scheduleID,
		Schedule: &models.Schedule{
			UUID: job.ScheduleUUID,
		},
		Logs:      []*models.Log{},
		CreatedAt: job.CreatedAt,
	}

	if scheduleID != nil {
		jobModel.Schedule.ID = *scheduleID
	}

	if job.Transaction != nil {
		jobModel.Transaction = NewTransactionModelFromEntities(job.Transaction)
	}

	for _, log := range job.Logs {
		jobModel.Logs = append(jobModel.Logs, NewLogModelFromEntity(log))
	}

	return jobModel
}

func NewJobEntityFromModels(jobModel *models.Job) *types.Job {
	job := &types.Job{
		UUID:        jobModel.UUID,
		ChainUUID:   jobModel.ChainUUID,
		Type:        jobModel.Type,
		Labels:      jobModel.Labels,
		CreatedAt:   jobModel.CreatedAt,
		Annotations: jobModel.Annotations,
		Logs:        []*types.Log{},
	}

	if jobModel.Schedule != nil {
		job.ScheduleUUID = jobModel.Schedule.UUID
	}

	if jobModel.Transaction != nil {
		job.Transaction = NewTransactionEntityFromModels(jobModel.Transaction)
	}

	for _, logModel := range jobModel.Logs {
		job.Logs = append(job.Logs, NewLogEntityFromModels(logModel))
	}

	return job
}

func NewEnvelopeFromJobModel(job *models.Job, headers map[string]string) *tx.TxEnvelope {
	contextLabels := job.Labels
	if contextLabels == nil {
		contextLabels = map[string]string{}
	}
	contextLabels["scheduleUUID"] = job.Schedule.UUID
	contextLabels["priority"] = job.Annotations.GasPricePolicy.Priority

	txEnvelope := &tx.TxEnvelope{
		Msg: &tx.TxEnvelope_TxRequest{TxRequest: &tx.TxRequest{
			Id:      job.UUID,
			Headers: headers,
			Params: &tx.Params{
				From:           job.Transaction.Sender,
				To:             job.Transaction.Recipient,
				Gas:            job.Transaction.Gas,
				GasPrice:       job.Transaction.GasPrice,
				Value:          job.Transaction.Value,
				Nonce:          job.Transaction.Nonce,
				Data:           job.Transaction.Data,
				Raw:            job.Transaction.Raw,
				PrivateFor:     job.Transaction.PrivateFor,
				PrivateFrom:    job.Transaction.PrivateFrom,
				PrivacyGroupId: job.Transaction.PrivacyGroupID,
			},
			ContextLabels: contextLabels,
			JobType:       tx.JobTypeMap[job.Type],
		}},
		InternalLabels: make(map[string]string),
	}

	txEnvelope.SetChainUUID(job.ChainUUID)

	chainID := new(big.Int)
	chainID.SetString(job.Annotations.ChainID, 10)
	txEnvelope.SetChainID(chainID)

	if job.Annotations.OneTimeKey {
		txEnvelope.EnableTxFromOneTimeKey()
	}

	return txEnvelope
}
