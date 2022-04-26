package chains

import (
	"github.com/consensys/orchestrate/pkg/utils"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type Block struct {
	chainUUID string
	Number    uint64
	TxHashes  []*ethcommon.Hash
}

func NewEthereumBlock(chainUUID string, b *ethtypes.Block) *Block {
	res := &Block{}
	res.chainUUID = chainUUID
	res.Number = b.Number().Uint64()
	res.TxHashes = []*ethcommon.Hash{}
	for _, tx := range b.Transactions() {
		res.TxHashes = append(res.TxHashes, utils.ToPtr(tx.Hash()).(*ethcommon.Hash))
	}
	return res
}
