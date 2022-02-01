// +build unit

package parsers

import (
	"testing"

	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/stretchr/testify/assert"
)

func TestChainsParser(t *testing.T) {
	chain := testdata.FakeChain()
	chainModel := NewChainModel(chain)
	finalChain := NewChainEntity(chainModel)

	assert.Equal(t, chain, finalChain)
}
