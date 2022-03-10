package types

import (
	"time"
)

type RegisterChainRequest struct {
	Name                string                  `json:"name" validate:"required" example:"mainnet"`                                                                                             // Name of the chain. Must be unique.
	URLs                []string                `json:"urls" pg:"urls,array" validate:"required,min=1,unique,dive,url" example:"https://mainnet.infura.io/v3/a73136601e6f4924a0baa4ed880b535e"` // List of URLs of Ethereum nodes to connect to.
	PrivateTxManagerURL string                  `json:"privateTxManagerURL,omitempty" validate:"omitempty,url"  example:"http://go-quorum/tessera:9000"`                                        // Private tx manager required by go-quorum for sending private txs
	Listener            RegisterListenerRequest `json:"listener,omitempty"`
	Labels              map[string]string       `json:"labels,omitempty"` // List of custom labels. Useful for adding custom information to the chain.
}

type RegisterListenerRequest struct {
	Depth             uint64 `json:"depth,omitempty" example:"0"`                                            // Block depth after which the Transaction Listener considers a block final and processes it (default 0).
	FromBlock         string `json:"fromBlock,omitempty" example:"latest"`                                   // Block from which the Transaction Listener should start processing transactions (default `latest`).
	BackOffDuration   string `json:"backOffDuration,omitempty" validate:"omitempty,isDuration" example:"1s"` // Time to wait before trying to fetch a new mined block (for example `1s` or `1m`, default is `5s`).
	ExternalTxEnabled bool   `json:"externalTxEnabled,omitempty" example:"false"`                            // Whether to listen to external transactions not crafted by Orchestrate (default `false`).
}

type UpdateChainRequest struct {
	Name     string                 `json:"name,omitempty" example:"mainnet"` // Name of the chain. Must be unique.
	Listener *UpdateListenerRequest `json:"listener,omitempty"`
	Labels   map[string]string      `json:"labels,omitempty"` // List of custom labels. Useful for adding custom information to the chain.
}

type UpdateListenerRequest struct {
	Depth             uint64 `json:"depth,omitempty" example:"0"`                                            // Block depth after which the Transaction Listener considers a block final and processes it (default 0).
	BackOffDuration   string `json:"backOffDuration,omitempty" validate:"omitempty,isDuration" example:"1s"` // Time to wait before trying to fetch a new mined block (for example `1s` or `1m`, default is `5s`).
	ExternalTxEnabled bool   `json:"externalTxEnabled,omitempty" example:"false"`                            // Whether to listen to external transactions not crafted by Orchestrate (default `false`).
	CurrentBlock      uint64 `json:"currentBlock,omitempty" example:"1"`                                     // Latest block number fetched.
}

type ChainResponse struct {
	UUID                      string            `json:"uuid" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`                                   // UUID of the registered chain.
	Name                      string            `json:"name" example:"mainnet"`                                                                // Name of the chain.
	TenantID                  string            `json:"tenantID" example:"tenant"`                                                             // ID of the tenant executing the API.
	OwnerID                   string            `json:"ownerID,omitempty" example:"foo"`                                                       // ID of the chain owner.
	URLs                      []string          `json:"urls" example:"https://mainnet.infura.io/v3/a73136601e6f4924a0baa4ed880b535e"`          // URLs of Ethereum nodes connected to.
	PrivateTxManagerURL       string            `json:"privateTxManagerURL,omitempty" validate:"url"  example:"http://go-quorum/tessera:9000"` // Private tx manager required by go-quorum for sending private txs
	ChainID                   string            `json:"chainID" example:"2445"`                                                                // [Ethereum chain ID](https://besu.hyperledger.org/en/latest/Concepts/NetworkID-And-ChainID/).
	ListenerDepth             uint64            `json:"listenerDepth" example:"0"`                                                             // Block depth after which the Transaction Listener considers a block final and processes it.
	ListenerCurrentBlock      uint64            `json:"listenerCurrentBlock" example:"0"`                                                      // Current block.
	ListenerStartingBlock     uint64            `json:"listenerStartingBlock" example:"5000"`                                                  // Block at which the Transaction Listener starts processing transactions
	ListenerBackOffDuration   string            `json:"listenerBackOffDuration" example:"5s"`                                                  // Time to wait before trying to fetch a new mined block.
	ListenerExternalTxEnabled bool              `json:"listenerExternalTxEnabled" example:"false"`                                             // Whether the chain listens for external transactions not crafted by Orchestrate.
	Labels                    map[string]string `json:"labels,omitempty"`                                                                      // List of custom labels.
	CreatedAt                 time.Time         `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`                                       // Date and time at which the chain was registered.
	UpdatedAt                 time.Time         `json:"updatedAt" example:"2020-07-09T12:35:42.115395Z"`                                       // Date and time at which the chain details were updated.
}
