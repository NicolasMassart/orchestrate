package testdata

import (
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/gofrs/uuid"
)

func FakeRegisterChainRequest() *api.RegisterChainRequest {
	return &api.RegisterChainRequest{
		Name: "mainnet",
		URLs: []string{"http://chain:8545"},
		Listener: api.RegisterListenerRequest{
			FromBlock:         "latest",
			ExternalTxEnabled: false,
		},
		PrivateTxManagerURL: "http://tessera-eea:8545",
		Labels: map[string]string{
			"label1": "val1",
			"label2": "val2",
		},
	}
}

func FakeUpdateChainRequest() *api.UpdateChainRequest {
	return &api.UpdateChainRequest{
		Name: "mainnet",
		Listener: &api.UpdateListenerRequest{
			CurrentBlock: 55,
		},
		Labels: map[string]string{
			"label3": "val3",
		},
	}
}

func FakeChainResponse() *api.ChainResponse {
	return &api.ChainResponse{
		UUID:                      uuid.Must(uuid.NewV4()).String(),
		Name:                      "ganache",
		TenantID:                  multitenancy.DefaultTenant,
		URLs:                      []string{"http://ethereum-node:8545"},
		ChainID:                   "888",
		ListenerDepth:             0,
		ListenerCurrentBlock:      0,
		ListenerStartingBlock:     0,
		ListenerBackOffDuration:   "5s",
		ListenerExternalTxEnabled: false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
}