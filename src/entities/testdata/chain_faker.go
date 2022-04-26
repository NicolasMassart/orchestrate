package testdata

import (
	"math/big"
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
		ChainID:                   big.NewInt(888),
		ListenerDepth:             0,
		ListenerBlockTimeDuration: blockTimeDuration,
	}
}

