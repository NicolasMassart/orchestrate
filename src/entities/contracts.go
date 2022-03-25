package entities

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const DefaultTagValue = "latest"

type Contract struct {
	Name             string
	Tag              string
	ABI              abi.ABI
	RawABI           string
	Bytecode         hexutil.Bytes
	DeployedBytecode hexutil.Bytes
	CodeHash         hexutil.Bytes
	Constructor      ABIComponent
	Methods          []ABIComponent
	Events           []ABIComponent
}

type ABIComponent struct {
	Signature string
	ABI       string
}

type Arguments struct {
	Name    string
	Type    string
	Indexed bool
}

type ContractEvent struct {
	ABI               string
	CodeHash          hexutil.Bytes
	SigHash           hexutil.Bytes
	IndexedInputCount uint
}

type Artifact struct {
	ABI              string
	Bytecode         hexutil.Bytes
	DeployedBytecode hexutil.Bytes
	CodeHash         hexutil.Bytes
}

type RawABI struct {
	Type      string
	Name      string
	Constant  bool
	Anonymous bool
	Inputs    []Arguments
	Outputs   []Arguments
}

func (c *Contract) String() string {
	tag := DefaultTagValue
	if c.Tag != "" {
		tag = c.Tag
	}

	return fmt.Sprintf("%v[%v]", c.Name, tag)
}
