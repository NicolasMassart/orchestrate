package events

import (
	"github.com/consensys/orchestrate/src/entities"
)

type ChainEventType string

const (
	NewChain     ChainEventType = "CREATED"
	UpdatedChain ChainEventType = "UPDATED"
	DeletedChain ChainEventType = "DELETED"
)

type Chain struct {
	Type  ChainEventType
	Chain *entities.Chain
}
