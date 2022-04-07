package utils

import (
	"context"
	"net/http"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/api/service/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

func ImportOrFetchAccount(ctx context.Context, sdkClient client.OrchestrateClient, address string, req *types.ImportAccountRequest) (*types.AccountResponse, error) {
	faucetAccRes, err := sdkClient.ImportAccount(ctx, req)
	if err == nil {
		return faucetAccRes, nil
	}

	if httpErr, ok := err.(*client.HTTPErr); !ok || httpErr.Code() != http.StatusConflict {
		return nil, err
	}

	return sdkClient.GetAccount(ctx, ethcommon.HexToAddress(address))
}
