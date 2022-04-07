package testdata

import (
	"fmt"

	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/quorum-key-manager/pkg/common"
)

func FakeRegisterChainRequest() *api.RegisterChainRequest {
	return &api.RegisterChainRequest{
		Name: "chain-" + common.RandString(5),
		URLs: []string{"http://chain:8545"},
		Listener: api.RegisterListenerRequest{
			FromBlock:         "latest",
			ExternalTxEnabled: false,
			Depth:             uint64(common.RandInt(10)),
			BackOffDuration:   fmt.Sprintf("%ds", common.RandIntRange(1, 10)),
		},
		PrivateTxManagerURL: "http://tessera-eea:8545",
		Labels: map[string]string{
			"label1": common.RandString(5),
			"label2": common.RandString(5),
		},
	}
}

func FakeUpdateChainRequest() *api.UpdateChainRequest {
	return &api.UpdateChainRequest{
		Name: "chain-" + common.RandString(5),
		Listener: &api.UpdateListenerRequest{
			CurrentBlock:      55,
			ExternalTxEnabled: true,
			Depth:             uint64(common.RandInt(10)),
			BackOffDuration:   fmt.Sprintf("%ds", common.RandIntRange(1, 10)),
		},
		Labels: map[string]string{
			"label3": common.RandString(5),
		},
	}
}
