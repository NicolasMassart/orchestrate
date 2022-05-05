package service

import (
	"github.com/consensys/orchestrate/src/infra/messenger"
)

var ContractEventMessageType messenger.ConsumerRequestMessageType = "contract_event"
var TransactionMessageType messenger.ConsumerRequestMessageType = "transaction_notification"
