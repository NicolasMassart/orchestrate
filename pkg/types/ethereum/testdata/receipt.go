package testdata

import (
	proto "github.com/consensys/orchestrate/pkg/types/ethereum"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/quorum-key-manager/pkg/common"
)

func FakeReceipt() *proto.Receipt {
	return &proto.Receipt{
		TxHash:      testdata.FakeTxHash().String(),
		BlockHash:   testdata.FakeHash().String(),
		Bloom:       testdata.FakeHash().String(),
		TxIndex:     uint64(common.RandInt(100)),
		BlockNumber: uint64(common.RandInt(100)),
		Status:      1,
		Logs:        []*proto.Log{FakeReceiptLogs()},
		GasUsed:     uint64(common.RandInt(1000000)),
	}
}

func FakeReceiptLogs() *proto.Log {
	return &proto.Log{
		Topics: []string{
			testdata.FakeHash().String(),
			testdata.FakeHash().String(),
		},
		Data:        testdata.FakeHash().String(),
		BlockNumber: 2019236,
		TxHash:      testdata.FakeHash().String(),
		TxIndex:     uint64(common.RandInt(100)),
		BlockHash:   testdata.FakeHash().String(),
		Removed:     true,
	}
}
