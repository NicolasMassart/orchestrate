package testdata

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"
)

func FakeChain() *entities.Chain {
	blockTimeDuration, _ := time.ParseDuration("5s")
	return &entities.Chain{
		UUID:                      uuid.Must(uuid.NewV4()).String(),
		Name:                      "ganache",
		TenantID:                  multitenancy.DefaultTenant,
		URLs:                      []string{"http://ethereum-node:8545"},
		ChainID:                   new(big.Int).SetUint64(rand.Uint64()),
		ListenerDepth:             0,
		ListenerBlockTimeDuration: blockTimeDuration,
	}
}

