package utils

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func WaitForBalance(ctx context.Context, ethClient *rpc.Client, chainProxyURL, address string) (*big.Int, error) {
	var balance *big.Int
	var gerr error
	err := backoff.Retry(func() error {
		balance, gerr = ethClient.BalanceAt(ctx, chainProxyURL, ethcommon.HexToAddress(address), nil)
		if gerr != nil {
			return backoff.Permanent(gerr)
		}
		if balance.Uint64() == 0 {
			return fmt.Errorf("%s has empty balance", address)
		}

		return nil
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 10))

	return balance, err
}
