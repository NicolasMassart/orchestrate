package testdata

import (
	"math/big"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/quorum-key-manager/pkg/common"
	"github.com/gofrs/uuid"
)

func FakeChain() *entities.Chain {
	backOffDuration, _ := time.ParseDuration("5s")
	return &entities.Chain{
		UUID:                      uuid.Must(uuid.NewV4()).String(),
		Name:                      "ganache",
		TenantID:                  multitenancy.DefaultTenant,
		URLs:                      []string{"http://ethereum-node:8545"},
		ChainID:                   big.NewInt(888),
		ListenerDepth:             0,
		ListenerCurrentBlock:      uint64(common.RandInt(100)),
		ListenerStartingBlock:     0,
		ListenerBackOffDuration:   backOffDuration,
		ListenerExternalTxEnabled: false,
		PrivateTxManager:          FakePrivateTxManager(),
	}
}

func FakePrivateTxManager() *entities.PrivateTxManager {
	return &entities.PrivateTxManager{
		UUID:      uuid.Must(uuid.NewV4()).String(),
		ChainUUID: uuid.Must(uuid.NewV4()).String(),
		URL:       "http://tessera:8545",
		Type:      "Tessera",
	}
}
