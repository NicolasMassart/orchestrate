package types

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// PrivateArgs are transaction arguments to provide to an Ethereum client supporting privacy (such as Quorum)
type PrivateArgs struct {
	// Private Transaction Fields
	PrivateFrom    []byte   `json:"privateFrom"`
	PrivateFor     [][]byte `json:"privateFor"`
	PrivacyGroupID []byte   `json:"privacyGroupId"`
	PrivateTxType  string   `json:"restriction"`
}

// SendTxArgs are arguments to provide to jsonRPC call `eth_sendTransaction`
type SendTxArgs struct {
	// From address in case of a non raw transaction
	From ethcommon.Address `json:"from"`

	// Main transaction attributes
	To       *ethcommon.Address `json:"to"`
	Gas      *hexutil.Uint64    `json:"gas"`
	GasPrice *hexutil.Big       `json:"gasPrice"`
	Value    *hexutil.Big       `json:"value"`
	Nonce    *hexutil.Uint64    `json:"nonce"`

	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred
	Data  hexutil.Bytes `json:"data"`
	Input hexutil.Bytes `json:"input"`

	// Private field
	PrivateArgs
}
