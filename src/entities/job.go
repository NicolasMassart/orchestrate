package entities

import (
	"time"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
)

type JobType string
type JobStatus string

const (
	EthereumTransaction       JobType = "eth://ethereum/transaction"       // Classic public Ethereum transaction
	EthereumRawTransaction    JobType = "eth://ethereum/rawTransaction"    // Classic raw transaction
	EEAMarkingTransaction     JobType = "eth://eea/markingTransaction"     // Besu marking transaction
	EEAPrivateTransaction     JobType = "eth://eea/privateTransaction"     // Besu private EEA tx
	TesseraMarkingTransaction JobType = "eth://tessera/markingTransaction" // Tessera public transaction
	TesseraPrivateTransaction JobType = "eth://tessera/privateTransaction" // Tessera private transaction
)

const (
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
