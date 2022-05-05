package service

import (
	"github.com/consensys/orchestrate/src/infra/messenger"
)

var PendingJobMessageType messenger.ConsumerRequestMessageType = "pending-job"
var SubscriptionMessageType messenger.ConsumerRequestMessageType = "subscription"
