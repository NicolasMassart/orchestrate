// +build unit

package parsers

import (
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/src/entities/testdata"

	"github.com/stretchr/testify/assert"
	modelstestdata "github.com/consensys/orchestrate/src/api/store/models/testdata"

	"encoding/json"
)

func TestParsersJob_NewModelFromEntity(t *testing.T) {
	jobEntity := testdata.FakeJob()
	finalJobEntity := NewJobEntity(NewJobModelFromEntities(jobEntity))

	expectedJSON, _ := json.Marshal(jobEntity)
	actualJSON, _ := json.Marshal(finalJobEntity)
	assert.Equal(t, string(expectedJSON), string(actualJSON))
}

func TestParsersJob_NewEntityFromModel(t *testing.T) {
	jobModel := modelstestdata.FakeJobModel(1)
	jobEntity := NewJobEntity(jobModel)
	finalJobModel := NewJobModelFromEntities(jobEntity)
	finalJobModel.ScheduleID = jobModel.ScheduleID
	finalJobModel.Schedule = jobModel.Schedule

	assert.Equal(t, finalJobModel.ScheduleID, jobModel.ScheduleID)
	assert.Equal(t, finalJobModel.UUID, jobModel.UUID)
	assert.Equal(t, finalJobModel.NextJobUUID, jobModel.NextJobUUID)
	assert.Equal(t, finalJobModel.Type, jobModel.Type)
	assert.Equal(t, finalJobModel.Labels, jobModel.Labels)
	assert.Equal(t, finalJobModel.CreatedAt, jobModel.CreatedAt)
}

func TestParsersJob_NewEnvelopeFromModel(t *testing.T) {
	job := testdata.FakeJob()
	headers := map[string]string{
		"Authorization": "Bearer MyToken",
	}
	txEnvelope := job.TxEnvelope(headers)

	txRequest := txEnvelope.GetTxRequest()
	assert.Equal(t, job.ChainUUID, txEnvelope.GetChainUUID())
	assert.Equal(t, entities.JobTypeToEnvelopeType[job.Type], txRequest.GetJobType())
	assert.Equal(t, job.Transaction.From.String(), txRequest.Params.GetFrom())
	assert.Equal(t, job.Transaction.To.String(), txRequest.Params.GetTo())
	assert.Equal(t, job.Transaction.Data.String(), txRequest.Params.GetData())
	assert.Equal(t, fmt.Sprintf("%d", *job.Transaction.Nonce), txRequest.Params.GetNonce())
	assert.Equal(t, job.Transaction.Raw.String(), txRequest.Params.GetRaw())
	assert.Equal(t, job.Transaction.GasPrice.String(), txRequest.Params.GetGasPrice())
	assert.Equal(t, fmt.Sprintf("%d", *job.Transaction.Gas), txRequest.Params.GetGas())
	assert.Equal(t, job.Transaction.PrivateFor, txRequest.Params.GetPrivateFor())
	assert.Equal(t, job.Transaction.PrivateFrom, txRequest.Params.GetPrivateFrom())
	assert.Equal(t, job.Transaction.PrivacyGroupID, txRequest.Params.GetPrivacyGroupId())
	assert.Equal(t, job.InternalData.ChainID.String(), txEnvelope.GetChainID())
}

func TestParsersJob_NewEnvelopeFromModelOneTimeKey(t *testing.T) {
	job := testdata.FakeJob()
	job.InternalData = &entities.InternalData{
		OneTimeKey: true,
	}

	txEnvelope := job.TxEnvelope(map[string]string{})

	evlp, err := txEnvelope.Envelope()
	assert.NoError(t, err)
	assert.True(t, evlp.IsOneTimeKeySignature())
}
