package types

import (
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type RegisterContractRequest struct {
	ABI              interface{}   `json:"abi,omitempty" validate:"required"`                                                                              // ABI of the contract.
	Bytecode         hexutil.Bytes `json:"bytecode,omitempty" validate:"omitempty" example:"0x6080604052348015600f57600080f" swaggertype:"string"`         // Bytecode of the contract.
	DeployedBytecode hexutil.Bytes `json:"deployedBytecode,omitempty" validate:"omitempty" example:"0x6080604052348015600f57600080f" swaggertype:"string"` // Deployed bytecode of the contract.
	Name             string        `json:"name" validate:"required" example:"ERC20"`                                                                       // Name of the contract.
	Tag              string        `json:"tag,omitempty" example:"v1.0.0"`                                                                                 // Optional tag attached to the contract.
}

type ContractResponse struct {
	Name             string                  `json:"name" example:"ERC20"`                                                                                                          // Name of the contract.
	Tag              string                  `json:"tag" example:"v1.0.0"`                                                                                                          // Optional tag attached to the contract.
	Registry         string                  `json:"registry" example:"registry.consensys.net/orchestrate"`                                                                         // URL of the contract's registry.
	ABI              string                  `json:"abi" example:"[{anonymous: false, inputs: [{indexed: false, name: account, type: address}, name: MinterAdded, type: event}]}]"` // ABI of the contract.
	Bytecode         string                  `json:"bytecode,omitempty" example:"0x6080604052348015600f57600080f..." swaggertype:"string"`                                          // Bytecode of the contract.
	DeployedBytecode string                  `json:"deployedBytecode,omitempty" example:"0x6080604052348015600f57600080f..." swaggertype:"string"`                                  // Deployed bytecode of the contract.
	Constructor      ABIComponentResponse    `json:"constructor"`                                                                                                                   // Contract constructor.
	Methods          []entities.ABIComponent `json:"methods"`                                                                                                                       // List of contract methods.
	Events           []entities.ABIComponent `json:"events"`                                                                                                                        // List of contract events.
}

type ABIComponentResponse struct {
	Signature string `json:"signature" example:"transfer(address,uint256)"`
	ABI       string `json:"abi,omitempty" example:"[{anonymous: false, inputs: [{indexed: false, name: account, type: address}, name: MinterAdded, type: event}]}]"`
}

type GetContractEventsRequest struct {
	SigHash           hexutil.Bytes `json:"sig_hash" validate:"required" example:"0x6080604052348015600f57600080f" swaggertype:"string"`
	IndexedInputCount uint32        `json:"indexed_input_count" validate:"omitempty" example:"1"`
}

type GetContractEventsBySignHashResponse struct {
	Event         string   `json:"event" validate:"omitempty" example:"{anonymous:false,inputs:[{indexed:true,name:from,type:address},{indexed:true,name:to,type:address},{indexed:false,name:value,type:uint256}],name:Transfer,type:event}"`              // Contract event name.
	DefaultEvents []string `json:"defaultEvents" validate:"omitempty" example:"[{anonymous:false,inputs:[{indexed:true,name:from,type:address},{indexed:true,name:to,type:address},{indexed:false,name:value,type:uint256}],name:Transfer,type:event},..."` // Default contract event names.
}

type SetContractCodeHashRequest struct {
	CodeHash hexutil.Bytes `json:"code_hash" validate:"required" example:"0x6080604052348015600f57600080f" swaggertype:"string"` // Contract code hash to set.
}

type SearchContractRequest struct {
	CodeHash hexutil.Bytes      `json:"code_hash" validate:"required" example:"0x6080604052348015600f57600080f" swaggertype:"string"`          // Contract code hash.
	Address  *ethcommon.Address `json:"address" validate:"required" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"` // Contract address.
}
