package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type CreateSubscriptionRequest struct {
	Address      ethcommon.Address `json:"-"`
	Chain        string            `json:"chain" validate:"required" example:"mainnet"`
	EventStream  string            `json:"eventStream,omitempty" validate:"required" example:"myWeebhookStream"`
	ContractName string            `json:"contractName" validate:"required" example:"MyContract"` // Name of the contract.
	ContractTag  string            `json:"contractTag,omitempty" example:"v1.1.0"`
	FromBlock    *uint64           `json:"fromBlock,omitempty" example:"123"`
}

func (r *CreateSubscriptionRequest) ToEntity() *entities.Subscription {
	subscription := &entities.Subscription{
		ContractAddress: r.Address,
		ContractName:    r.ContractName,
		ContractTag:     r.ContractTag,
		FromBlock:       r.FromBlock,
	}

	if subscription.ContractTag == "" {
		subscription.ContractTag = entities.DefaultContractTagValue
	}

	return subscription
}

type UpdateSubscriptionRequest struct {
	EventStream string `json:"event_stream,omitempty" validate:"omitempty" example:"myWeebhookStream"`
}

func (r *UpdateSubscriptionRequest) ToEntity(uuid string) *entities.Subscription {
	return &entities.Subscription{
		UUID: uuid,
	}
}

type SubscriptionResponse struct {
	UUID         string    `json:"uuid,omitempty" validate:"omitempty"`
	ChainUUID    string    `json:"chainUUID,omitempty"  example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	ContractName string    `json:"contractName" validate:"required" example:"MyContract"` // Name of the contract.
	ContractTag  string    `json:"contractTag,omitempty" example:"v1.1.0"`
	EventStream  string    `json:"event_stream,omitempty" validate:"omitempty" example:"myWeebhookStream"`
	FromBlock    *uint64   `json:"fromBlock,omitempty" example:"123"`
	TenantID     string    `json:"tenantID"  example:"foo"`
	OwnerID      string    `json:"ownerID,omitempty"  example:"foo"`
	CreatedAt    time.Time `json:"createdAt"  example:"2020-07-09T12:35:42.115395Z"`
	UpdatedAt    time.Time `json:"updatedAt"  example:"2020-07-09T12:35:42.115395Z"`
}

func NewSubscriptionResponses(subs []*entities.Subscription) []*SubscriptionResponse {
	response := []*SubscriptionResponse{}
	for _, e := range subs {
		response = append(response, NewSubscriptionResponse(e))
	}

	return response
}

func NewSubscriptionResponse(sub *entities.Subscription) *SubscriptionResponse {
	return &SubscriptionResponse{
		UUID:         sub.UUID,
		ChainUUID:    sub.ChainUUID,
		ContractName: sub.ContractName,
		ContractTag:  sub.ContractTag,
		FromBlock:    sub.FromBlock,
		TenantID:     sub.TenantID,
		OwnerID:      sub.OwnerID,
		CreatedAt:    sub.CreatedAt,
		UpdatedAt:    sub.UpdatedAt,
	}
}
