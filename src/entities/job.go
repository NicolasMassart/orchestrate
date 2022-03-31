package entities

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
)

type JobType string
type JobStatus string

var (
	EthereumTransaction        JobType = "eth://ethereum/transaction"         // Classic public Ethereum transaction
	EthereumRawTransaction     JobType = "eth://ethereum/rawTransaction"      // Classic raw transaction
	EEAMarkingTransaction      JobType = "eth://eea/markingTransaction"       // Besu marking transaction
	EEAPrivateTransaction      JobType = "eth://eea/privateTransaction"       // Besu private EEA tx
	GoQuorumMarkingTransaction JobType = "eth://go-quorum/markingTransaction" // GoQuorum public transaction
	GoQuorumPrivateTransaction JobType = "eth://go-quorum/privateTransaction" // GoQuorum private transaction
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

func (job *Job) PartitionKey() string {
	// Return empty partition key for raw tx and one time key tx
	if job.Type == EthereumRawTransaction || job.InternalData.OneTimeKey {
		return ""
	}

	// @TODO Use chainID instead of chainUUID
	switch {
	case job.Type == EEAPrivateTransaction && job.Transaction.PrivacyGroupID != "":
		return fmt.Sprintf("%v@eea-%v@%v", job.Transaction.From, job.Transaction.PrivacyGroupID, job.ChainUUID)
	case job.Type == EEAPrivateTransaction && len(job.Transaction.PrivateFor) > 0:
		l := append(job.Transaction.PrivateFor, job.Transaction.PrivateFrom)
		lSorted := make([]string, len(l))
		copy(lSorted, l)
		sort.Strings(lSorted)
		h := md5.New()
		_, _ = h.Write([]byte(strings.Join(l, "-")))
		return fmt.Sprintf("%v@eea-%v@%v", job.Transaction.From, fmt.Sprintf("%x", h.Sum(nil)), job.ChainUUID)
	default:
		return fmt.Sprintf("%v@%v", job.Transaction.From, job.ChainUUID)
	}
}
