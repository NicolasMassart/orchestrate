package types

import (
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type SendTransactionRequest struct {
	ChainName string            `json:"chain" validate:"required" example:"myChain"` // Name of the chain on which to send the transaction.
	Labels    map[string]string `json:"labels,omitempty"`                            // List of custom labels.
	Params    TransactionParams `json:"params" validate:"required"`
}

// go validator does not support mutually exclusive parameters for now
// See more https://github.com/go-playground/validator/issues/608
type TransactionParams struct {
	Value           *hexutil.Big                  `json:"value,omitempty" validate:"omitempty" example:"0x44300E0" swaggertype:"string"`                                     // Value transferred, in Wei.
	Gas             *uint64                       `json:"gas,omitempty" example:"50000"`                                                                                     // Gas provided by the sender.
	GasPrice        *hexutil.Big                  `json:"gasPrice,omitempty" validate:"omitempty" example:"0xAB208" swaggertype:"string"`                                    // If sending a non-[EIP1559 transaction](https://besu.hyperledger.org/en/stable/Concepts/Transactions/Transaction-Types/#eip1559-transactions), the gas price, in Wei, provided by the sender.
	GasFeeCap       *hexutil.Big                  `json:"maxFeePerGas,omitempty" example:"0x4c4b40" swaggertype:"string"`                                                    // If sending an [EIP1559 transaction](https://besu.hyperledger.org/en/stable/Concepts/Transactions/Transaction-Types/#eip1559-transactions), the maximum total fee, in Wei, the sender is willing to pay per gas.
	GasTipCap       *hexutil.Big                  `json:"maxPriorityFeePerGas,omitempty" example:"0x59682f00" swaggertype:"string"`                                          // If sending an [EIP1559 transaction](https://besu.hyperledger.org/en/stable/Concepts/Transactions/Transaction-Types/#eip1559-transactions), the maximum fee, in Wei, the sender is willing to pay per gas above the base fee.
	AccessList      types.AccessList              `json:"accessList,omitempty" swaggertype:"array,object"`                                                                   // Optional list of addresses and storage keys the transaction plans to access.
	TransactionType string                        `json:"transactionType,omitempty" validate:"omitempty,isTransactionType" example:"dynamic_fee" enums:"legacy,dynamic_fee"` // `dynamic_fee` for a post-London fork transaction, `legacy` for a pre-London fork transaction.
	From            *ethcommon.Address            `json:"from" validate:"omitempty" example:"0x1abae27a0cbfb02945720425d3b80c7e097285534" swaggertype:"string"`              // Address of the sender.
	To              *ethcommon.Address            `json:"to" validate:"required" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"`                  // Address of the recipient, mutually exclusive with `from`.
	MethodSignature string                        `json:"methodSignature" validate:"required" example:"transfer(address,uint256)"`
	Args            []interface{}                 `json:"args,omitempty"`                      // Contract arguments.
	OneTimeKey      bool                          `json:"oneTimeKey,omitempty" example:"true"` // Indicates if the transaction is a One Time Key transaction.
	GasPricePolicy  GasPriceParams                `json:"gasPricePolicy,omitempty"`
	Protocol        entities.PrivateTxManagerType `json:"protocol,omitempty" validate:"omitempty,isPrivateTxManagerType" example:"Tessera"`                                                                                           // Currently supports `Tessera` and `EEA`.
	PrivateFrom     string                        `json:"privateFrom,omitempty" validate:"omitempty,base64" example:"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="`                                                                   // When sending a private transaction, the sender's public key.
	PrivateFor      []string                      `json:"privateFor,omitempty" validate:"omitempty,min=1,unique,dive,base64" example:"[A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=,B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=]"`   // When sending a private transaction, a list of the recipients' public keys. Not used with `PrivacyGroupID`.
	MandatoryFor    []string                      `json:"mandatoryFor,omitempty" validate:"omitempty,min=1,unique,dive,base64" example:"[A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=,B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=]"` // When sending a private transaction with [mandatory party protection](https://consensys.net/docs/goquorum/en/stable/concepts/privacy/privacy-enhancements/#mandatory-party-protection), a list of the recipients' public keys.
	PrivacyGroupID  string                        `json:"privacyGroupId,omitempty" validate:"omitempty,base64" example:"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="`                                                                // When sending a private transaction, the privacy group ID of the recipients. Not used with `PrivateFor`.
	PrivacyFlag     entities.PrivacyFlag          `json:"privacyFlag,omitempty" validate:"omitempty,isPrivacyFlag" example:"0"`                                                                                                       // Set to 0 for standard privacy (default), 1 for [counter-party protection](https://consensys.net/docs/goquorum/en/stable/concepts/privacy/privacy-enhancements/#counter-party-protection), 2 for [mandatory party protection](https://consensys.net/docs/goquorum/en/stable/concepts/privacy/privacy-enhancements/#mandatory-party-protection), and 3 for [private state validation](https://consensys.net/docs/goquorum/en/latest/concepts/privacy/privacy-enhancements/#private-state-validation).
	ContractName    string                        `json:"contractName" validate:"required" example:"MyContract"`                                                                                                                      // Name of the contract.
	ContractTag     string                        `json:"contractTag,omitempty" example:"v1.1.0"`                                                                                                                                     // Optional tag attached to the contract.
}

func (params *TransactionParams) Validate() error {
	if err := infra.GetValidator().Struct(params); err != nil {
		return err
	}

	if params.Protocol != "" || params.PrivateFrom != "" {
		return validatePrivateTxParams(params.Protocol, params.PrivateFrom, params.PrivacyGroupID, params.PrivateFor)
	}

	if err := validateTxFromParams(params.From, params.OneTimeKey); err != nil {
		return err
	}

	return params.GasPricePolicy.RetryPolicy.Validate()
}