package testdata

import (
	"math/big"

	"github.com/consensys/orchestrate/src/entities"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

func FakeBlock(chainID uint64, blockNumber uint64, txs ...*entities.ETHTransaction) *ethtypes.Block {
	header := &ethtypes.Header{
		Number: new(big.Int).SetUint64(blockNumber),
	}
	ethTxs := []*ethtypes.Transaction{}
	for _, tx := range txs {
		ethTxs = append(ethTxs, tx.ToETHTransaction(new(big.Int).SetUint64(chainID)))
	}
	return ethtypes.NewBlock(header, ethTxs, []*ethtypes.Header{}, []*ethtypes.Receipt{}, new(trie.Trie))
}
