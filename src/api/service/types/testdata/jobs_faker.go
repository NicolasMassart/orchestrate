package testdata

import (
	"encoding/json"

	"github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/listener"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type minedMessageRequestMatcher struct {
	JobUUID string
	Status  entities.JobStatus
	Receipt *ethereum.Receipt
}

func MinedJobMessageRequestMatcher(JobUUID string, Receipt *ethereum.Receipt) *minedMessageRequestMatcher {
	return &minedMessageRequestMatcher{
		JobUUID: JobUUID,
		Status:  entities.StatusMined,
		Receipt: Receipt,
	}
}

func (j *minedMessageRequestMatcher) Matches(x interface{}) bool {
	xReq, ok := x.(*types.JobUpdateMessageRequest)
	if !ok {
		return false
	}
	if xReq.JobUUID != j.JobUUID {
		return false
	}
	if j.Status != "" && xReq.Status != j.Status {
		return false
	}
	if j.Receipt != nil && xReq.Receipt != j.Receipt {
		return false
	}
	return true
}

func (j *minedMessageRequestMatcher) String() string {
	return j.JobUUID
}

func NewFakeJobUpdateMessage(jobUUID string) *entities.Message {
	req := &types.JobUpdateMessageRequest{
		JobUUID: jobUUID,
	}
	bBody, _ := json.Marshal(req)
	return &entities.Message{
		Type: listener.UpdateJobMessageType,
		Body: bBody,
		Offset: int64(utils.RandInt(10)),
	}
}


type sentMessageRequestMatcher struct {
	JobUUID string
	Status  entities.JobStatus
	Hash    *ethcommon.Hash
}

func SentJobMessageRequestMatcher(JobUUID string, Status entities.JobStatus, hash *ethcommon.Hash) *sentMessageRequestMatcher {
	return &sentMessageRequestMatcher{
		JobUUID: JobUUID,
		Status:  Status,
		Hash:    hash,
	}
}

func (j *sentMessageRequestMatcher) Matches(x interface{}) bool {
	xReq, ok := x.(*types.JobUpdateMessageRequest)
	if !ok {
		return false
	}
	if xReq.JobUUID != j.JobUUID {
		return false
	}
	if j.Status != "" && xReq.Status != j.Status {
		return false
	}
	if j.Hash != nil {
		if xReq.Transaction.Hash == nil {
			return false
		}
		if xReq.Transaction.Hash.String() != j.Hash.String() {
			return false
		}
	}
	return true
}

func (j *sentMessageRequestMatcher) String() string {
	return j.JobUUID
}
