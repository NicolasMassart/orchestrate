package parsers

import (
	"testing"

	"github.com/ConsenSys/orchestrate/pkg/types/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAccountsParser(t *testing.T) {
	account := testutils.FakeAccount()
	accountModel := NewAccountModelFromEntities(account)
	finalAccount := NewAccountEntityFromModels(accountModel)

	assert.Equal(t, account, finalAccount)
}
