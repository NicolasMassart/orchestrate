// +build unit

package parsers

import (
	"github.com/consensys/orchestrate/src/entities/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
	modelstestdata "github.com/consensys/orchestrate/src/api/store/models/testdata"

	"encoding/json"
)

func TestParsersTransaction_NewModelFromEntity(t *testing.T) {
	txEntity := testdata.FakeETHTransaction()
	txModel := NewTransactionModel(txEntity)
	finalTxEntity := NewTransactionEntity(txModel)

	expectedJSON, _ := json.Marshal(txEntity)
	actualJOSN, _ := json.Marshal(finalTxEntity)
	assert.Equal(t, string(expectedJSON), string(actualJOSN))
}

func TestParsersTransaction_NewEntityFromModel(t *testing.T) {
	txModel := modelstestdata.FakeTransaction()
	txEntity := NewTransactionEntity(txModel)
	finalTxModel := NewTransactionModel(txEntity)
	finalTxModel.UUID = txModel.UUID

	expectedJSON, _ := json.Marshal(txModel)
	actualJOSN, _ := json.Marshal(finalTxModel)
	assert.Equal(t, string(expectedJSON), string(actualJOSN))
}

