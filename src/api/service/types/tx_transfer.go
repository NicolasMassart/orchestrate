package types

import (
	infra "github.com/consensys/orchestrate/src/infra/api"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransferRequest struct {
	ChainName string            `json:"chain" validate:"required" example:"myChain"` // Name of the chain on which to send the transaction.
	Labels    map[string]string `json:"labels,omitempty"`                            // List of custom labels.
	Params    TransferParams    `json:"params" validate:"required"`
}

type TransferParams struct {
	Value           *hexutil.Big      `json:"value" validate:"required" example:"0x59682f00" swaggertype:"string"`                                               // Value transferred, in Wei.
	Gas             *uint64           `json:"gas,omitempty" example:"21000"`                                                                                     // Gas provided by the sender.
	GasPrice        *hexutil.Big      `json:"gasPrice,omitempty" example:"0x5208" swaggertype:"string"`                                                          // If sending a non-[EIP1559 transaction](https://besu.hyperledger.org/en/stable/Concepts/Transactions/Transaction-Types/#eip1559-transactions), the gas price, in Wei, provided by the sender.
	GasFeeCap       *hexutil.Big      `json:"maxFeePerGas,omitempty" example:"0x4c4b40" swaggertype:"string"`                                                    // If sending an [EIP1559 transaction](https://besu.hyperledger.org/en/stable/Concepts/Transactions/Transaction-Types/#eip1559-transactions), the maximum total fee, in Wei, the sender is willing to pay per gas.
	GasTipCap       *hexutil.Big      `json:"maxPriorityFeePerGas,omitempty" example:"0x59682f00" swaggertype:"string"`                                          // If sending an [EIP1559 transaction](https://besu.hyperledger.org/en/stable/Concepts/Transactions/Transaction-Types/#eip1559-transactions), the maximum fee, in Wei, the sender is willing to pay per gas above the base fee.
	AccessList      types.AccessList  `json:"accessList,omitempty" swaggertype:"array,object"`                                                                   // Optional list of addresses and storage keys the transaction plans to access.
	TransactionType string            `json:"transactionType,omitempty" validate:"omitempty,isTransactionType" example:"dynamic_fee" enums:"legacy,dynamic_fee"` // `dynamic_fee` for a post-London fork transaction, `legacy` for a pre-London fork transaction.
	From            ethcommon.Address `json:"from" validate:"required" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"`                // Address of the sender.
	To              ethcommon.Address `json:"to" validate:"required" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"`                  // Address of the recipient, mutually exclusive with `from`.
	GasPricePolicy  GasPriceParams    `json:"gasPricePolicy,omitempty"`
}

func (params *TransferParams) Validate() error {
	if err := infra.GetValidator().Struct(params); err != nil {
		return err
	}
	return params.GasPricePolicy.RetryPolicy.Validate()
}
