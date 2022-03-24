package entities

import (
	"time"

	authutils "github.com/consensys/orchestrate/pkg/toolkit/app/auth/utils"
	ierror "github.com/consensys/orchestrate/pkg/types/error"
	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/types/tx"
	"github.com/consensys/orchestrate/pkg/utils"
)

type JobType string
type JobStatus string

var (
	EthereumTransaction       JobType = "eth://ethereum/transaction"         // Classic public Ethereum transaction
	EthereumRawTransaction    JobType = "eth://ethereum/rawTransaction"      // Classic raw transaction
	EEAMarkingTransaction     JobType = "eth://eea/markingTransaction"       // Besu marking transaction
	EEAPrivateTransaction     JobType = "eth://eea/privateTransaction"       // Besu private EEA tx
	TesseraMarkingTransaction JobType = "eth://go-quorum/markingTransaction" // GoQuorum public transaction
	TesseraPrivateTransaction JobType = "eth://go-quorum/privateTransaction" // GoQuorum private transaction
)

func (jt *JobType) String() string {
	return string(*jt)
}

var (
	StatusCreated    JobStatus = "CREATED"
	StatusStarted    JobStatus = "STARTED"
	StatusPending    JobStatus = "PENDING"
	StatusResending  JobStatus = "RESENDING"
	StatusStored     JobStatus = "STORED"
	StatusRecovering JobStatus = "RECOVERING"
	StatusWarning    JobStatus = "WARNING"
	StatusFailed     JobStatus = "FAILED"
	StatusMined      JobStatus = "MINED"
	StatusNeverMined JobStatus = "NEVER_MINED"
)

func (js *JobStatus) String() string {
	return string(*js)
}

type Job struct {
	UUID         string
	NextJobUUID  string
	ChainUUID    string
	ScheduleUUID string
	TenantID     string
	OwnerID      string
	Type         JobType
	Status       JobStatus
	Labels       map[string]string
	InternalData *InternalData
	Transaction  *ETHTransaction
	Receipt      *ethereum.Receipt
	Logs         []*Log
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func IsFinalJobStatus(status JobStatus) bool {
	return status == StatusMined ||
		status == StatusFailed ||
		status == StatusStored ||
		status == StatusNeverMined
}

func (job *Job) ShouldBeRetried() bool {
	if job.InternalData.ParentJobUUID != "" {
		return false
	}

	if job.InternalData.HasBeenRetried {
		return false
	}

	if job.InternalData.RetryInterval == 0 {
		return false
	}

	return true
}

func (job *Job) TxRequestEnvelope(headers map[string]string) *tx.TxEnvelope {
	contextLabels := job.Labels
	if contextLabels == nil {
		contextLabels = map[string]string{}
	}

	contextLabels[tx.NextJobUUIDLabel] = job.NextJobUUID
	contextLabels[tx.PriorityLabel] = job.InternalData.Priority
	contextLabels[tx.ParentJobUUIDLabel] = job.InternalData.ParentJobUUID

	txEnvelope := &tx.TxEnvelope{
		Msg: &tx.TxEnvelope_TxRequest{TxRequest: &tx.TxRequest{
			Id:      job.ScheduleUUID,
			Headers: headers,
			Params: &tx.Params{
				From:            utils.StringerToString(job.Transaction.From),
				To:              utils.StringerToString(job.Transaction.To),
				Gas:             utils.ValueToString(job.Transaction.Gas),
				GasPrice:        utils.StringerToString(job.Transaction.GasPrice),
				GasFeeCap:       utils.StringerToString(job.Transaction.GasFeeCap),
				GasTipCap:       utils.StringerToString(job.Transaction.GasTipCap),
				Value:           utils.StringerToString(job.Transaction.Value),
				Nonce:           utils.ValueToString(job.Transaction.Nonce),
				Data:            utils.StringerToString(job.Transaction.Data),
				Raw:             utils.StringerToString(job.Transaction.Raw),
				PrivateFrom:     job.Transaction.PrivateFrom,
				PrivateFor:      job.Transaction.PrivateFor,
				MandatoryFor:    job.Transaction.MandatoryFor,
				PrivacyGroupId:  job.Transaction.PrivacyGroupID,
				PrivacyFlag:     int32(job.Transaction.PrivacyFlag),
				TransactionType: string(job.Transaction.TransactionType),
				AccessList:      ConvertFromAccessList(job.Transaction.AccessList),
			},
			ContextLabels: contextLabels,
			JobType:       JobTypeToEnvelopeType[job.Type],
		}},
		InternalLabels: make(map[string]string),
	}

	txEnvelope.SetChainUUID(job.ChainUUID)

	if job.InternalData.ChainID != nil {
		txEnvelope.SetChainID(job.InternalData.ChainID)
	}

	txEnvelope.SetScheduleUUID(job.ScheduleUUID)
	txEnvelope.SetJobUUID(job.UUID)

	if job.InternalData.OneTimeKey {
		txEnvelope.EnableTxFromOneTimeKey()
	}

	if job.InternalData.ParentJobUUID != "" {
		txEnvelope.SetParentJobUUID(job.InternalData.ParentJobUUID)
	}

	if job.InternalData.Priority != "" {
		txEnvelope.SetPriority(job.InternalData.Priority)
	}

	if job.InternalData.StoreID != "" {
		txEnvelope.SetStoreID(job.InternalData.StoreID)
	}

	if job.Transaction.Hash != nil {
		txEnvelope.SetTxHash(job.Transaction.Hash.String())
	}

	return txEnvelope
}

func (job *Job) TxResponse(errs ...*ierror.Error) *tx.TxResponse {
	return &tx.TxResponse{
		Id:            job.ScheduleUUID,
		JobUUID:       job.UUID,
		ContextLabels: job.Labels,
		Transaction: &ethereum.Transaction{
			From:       utils.StringerToString(job.Transaction.From),
			Nonce:      utils.ValueToString(job.Transaction.Nonce),
			To:         utils.StringerToString(job.Transaction.To),
			Value:      utils.StringerToString(job.Transaction.Value),
			Gas:        utils.ValueToString(job.Transaction.Gas),
			GasPrice:   utils.StringerToString(job.Transaction.GasPrice),
			GasFeeCap:  utils.StringerToString(job.Transaction.GasFeeCap),
			GasTipCap:  utils.StringerToString(job.Transaction.GasTipCap),
			Data:       utils.StringerToString(job.Transaction.Data),
			Raw:        utils.StringerToString(job.Transaction.Raw),
			TxHash:     utils.StringerToString(job.Transaction.Hash),
			AccessList: ConvertFromAccessList(job.Transaction.AccessList),
			TxType:     string(job.Transaction.TransactionType),
		},
		Receipt: job.Receipt,
		Chain:   job.ChainUUID,
		Errors:  errs,
	}
}

func NewJobFromEnvelope(envelope *tx.Envelope) *Job {
	return &Job{
		UUID:         envelope.GetJobUUID(),
		NextJobUUID:  envelope.GetNextJobUUID(),
		ChainUUID:    envelope.GetChainUUID(),
		ScheduleUUID: envelope.GetScheduleUUID(),
		Type:         JobType(envelope.GetJobTypeString()),
		InternalData: &InternalData{
			OneTimeKey:    envelope.IsOneTimeKeySignature(),
			ChainID:       envelope.GetChainID(),
			ParentJobUUID: envelope.GetParentJobUUID(),
			Priority:      envelope.GetPriority(),
			StoreID:       envelope.GetStoreID(),
		},
		TenantID: envelope.GetHeadersValue(authutils.TenantIDHeader),
		OwnerID:  envelope.GetHeadersValue(authutils.UsernameHeader),
		Transaction: &ETHTransaction{
			Hash:            envelope.GetTxHash(),
			From:            envelope.GetFrom(),
			To:              envelope.GetTo(),
			Nonce:           envelope.GetNonce(),
			Value:           envelope.GetValue(),
			GasPrice:        envelope.GetGasPrice(),
			Gas:             envelope.GetGas(),
			GasFeeCap:       envelope.GetGasFeeCap(),
			GasTipCap:       envelope.GetGasTipCap(),
			AccessList:      ConvertToAccessList(envelope.GetAccessList()),
			TransactionType: TransactionType(envelope.GetTransactionType()),
			Data:            envelope.GetData(),
			Raw:             envelope.GetRaw(),
			PrivateFrom:     envelope.GetPrivateFrom(),
			PrivateFor:      envelope.GetPrivateFor(),
			MandatoryFor:    envelope.GetMandatoryFor(),
			PrivacyGroupID:  envelope.GetPrivacyGroupID(),
			PrivacyFlag:     PrivacyFlag(envelope.GetPrivacyFlag()),
			EnclaveKey:      utils.StringToHexBytes(envelope.GetEnclaveKey()),
		},
	}
}

var JobTypeToEnvelopeType = map[JobType]tx.JobType{
	EthereumTransaction:       tx.JobType_ETH_TX,
	EthereumRawTransaction:    tx.JobType_ETH_RAW_TX,
	EEAMarkingTransaction:     tx.JobType_EEA_MARKING_TX,
	EEAPrivateTransaction:     tx.JobType_EEA_PRIVATE_TX,
	TesseraMarkingTransaction: tx.JobType_GO_QUORUM_MARKING_TX,
	TesseraPrivateTransaction: tx.JobType_GO_QUORUM_PRIVATE_TX,
}
