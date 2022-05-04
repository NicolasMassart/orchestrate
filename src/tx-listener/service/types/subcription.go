package types

import (
	"github.com/consensys/orchestrate/src/entities"
)

type SubscriptionAction string

var CreateSubscriptionAction SubscriptionAction = "CREATED"
var UpdateSubscriptionAction SubscriptionAction = "UPDATE"
var DeleteSubscriptionAction SubscriptionAction = "DELETED"

// @TODO Make it based on chainID
type SubscriptionMessageRequest struct {
	Subscription *entities.Subscription
	Action       SubscriptionAction
}
