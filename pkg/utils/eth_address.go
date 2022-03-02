package utils

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

const ZeroAddressString = "0x0000000000000000000000000000000000000000"

func ParseHexToMixedCaseEthAddress(address string) (*ethcommon.Address, error) {
	if !ethcommon.IsHexAddress(address) {
		return nil, fmt.Errorf("expected hex string")
	}

	addr := ethcommon.HexToAddress(address)
	return &addr, nil
}
