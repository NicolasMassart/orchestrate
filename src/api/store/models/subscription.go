package models

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type Subscription struct {
	tableName struct{} `pg:"subscriptions"` // nolint:unused,structcheck // reason

	ID              int `pg:"alias:id"`
	UUID            string
	ContractAddress string
	ContractName    string
	ContractTag     string
	EventStreamUUID string    `pg:"alias:event_stream_uuid"`
	ChainUUID       string    `pg:"alias:chain_uuid"`
	TenantID        string    `pg:"alias:tenant_id"`
	OwnerID         string    `pg:"alias:owner_id"`
	CreatedAt       time.Time `pg:"default:now()"`
	UpdatedAt       time.Time `pg:"default:now()"`
}

func NewSubscription(sub *entities.Subscription) *Subscription {
	es := &Subscription{
		UUID:            sub.UUID,
		ContractAddress: sub.ContractAddress.String(),
		ContractName:    sub.ContractName,
		ContractTag:     sub.ContractTag,
		ChainUUID:       sub.ChainUUID,
		EventStreamUUID: sub.EventStreamUUID,
		TenantID:        sub.TenantID,
		OwnerID:         sub.OwnerID,
		CreatedAt:       sub.CreatedAt,
		UpdatedAt:       sub.UpdatedAt,
	}

	return es
}

func NewSubscriptions(subs []*Subscription) []*entities.Subscription {
	res := []*entities.Subscription{}
	for _, e := range subs {
		res = append(res, e.ToEntity())
	}

	return res
}

func (e *Subscription) ToEntity() *entities.Subscription {
	es := &entities.Subscription{
		UUID:            e.UUID,
		ContractAddress: ethcommon.HexToAddress(e.ContractAddress),
		ContractName:    e.ContractName,
		ContractTag:     e.ContractTag,
		EventStreamUUID: e.EventStreamUUID,
		ChainUUID:       e.ChainUUID,
		TenantID:        e.TenantID,
		OwnerID:         e.OwnerID,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}

	return es
}
