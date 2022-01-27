package testdata

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
)

func FakeCreateAccountRequest() *api.CreateAccountRequest {
	return &api.CreateAccountRequest{
		Alias: fmt.Sprintf("Alias_%s", utils.RandString(10)),
		Attributes: map[string]string{
			"attr1": "val1",
			"attr2": "val2",
		},
	}
}

func FakeImportAccountRequest() *api.ImportAccountRequest {
	return &api.ImportAccountRequest{
		Alias:      fmt.Sprintf("Alias_%s", utils.RandString(10)),
		PrivateKey: hexutil.MustDecode("0xa93e498896143c02fdf42b9b69bdcf4aebcedc8d45851c33f8ae86057e7c4a90"),
		Attributes: map[string]string{
			"attr1": "val1",
			"attr2": "val2",
		},
	}
}

func FakeUpdateAccountRequest() *api.UpdateAccountRequest {
	return &api.UpdateAccountRequest{
		Alias: fmt.Sprintf("Alias_%s", utils.RandString(5)),
		Attributes: map[string]string{
			"attr3": "val3",
			"attr4": "val4",
		},
	}
}
