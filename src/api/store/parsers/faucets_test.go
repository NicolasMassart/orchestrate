// +build unit

package parsers

import (
	"testing"

	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/stretchr/testify/assert"
)

func TestFaucetsParser(t *testing.T) {
	faucet := testdata.FakeFaucet()
	faucetModel := NewFaucetModel(faucet)
	finalFaucet := NewFaucetEntity(faucetModel)

	assert.Equal(t, faucet, finalFaucet)
}
