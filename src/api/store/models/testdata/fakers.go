package testdata

import (
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/utils"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/gofrs/uuid"
)

func FakeSchedule(tenantID, username string) *models.Schedule {
	if tenantID == "" {
		tenantID = multitenancy.DefaultTenant
	}
	return &models.Schedule{
		TenantID: tenantID,
		OwnerID:  username,
		UUID:     uuid.Must(uuid.NewV4()).String(),
		Jobs: []*models.Job{{
			UUID:        uuid.Must(uuid.NewV4()).String(),
			ChainUUID:   uuid.Must(uuid.NewV4()).String(),
			Type:        entities.EthereumTransaction.String(),
			Transaction: FakeTransaction(),
			Logs:        []*models.Log{{Status: entities.StatusCreated.String(), Message: "created message"}},
		}},
	}
}

func FakeTransaction() *models.Transaction {
	return &models.Transaction{
		UUID: uuid.Must(uuid.NewV4()).String(),
	}
}

func FakeJobModel(scheduleID int) *models.Job {
	job := &models.Job{
		UUID:      uuid.Must(uuid.NewV4()).String(),
		ChainUUID: uuid.Must(uuid.NewV4()).String(),
		Type:      entities.EthereumTransaction.String(),
		Status:    entities.StatusCreated.String(),
		Schedule: &models.Schedule{
			ID:       scheduleID,
			TenantID: multitenancy.DefaultTenant,
			UUID:     uuid.Must(uuid.NewV4()).String(),
		},
		Transaction: FakeTransaction(),
		Logs: []*models.Log{
			{UUID: uuid.Must(uuid.NewV4()).String(), Status: entities.StatusCreated.String(), Message: "created message", CreatedAt: time.Now()},
		},
		InternalData: &entities.InternalData{
			ChainID: big.NewInt(888),
		},
		CreatedAt: time.Now(),
		Labels:    make(map[string]string),
	}

	if scheduleID != 0 {
		job.ScheduleID = &scheduleID
	}

	return job
}

func FakeLog() *models.Log {
	return &models.Log{
		UUID:      uuid.Must(uuid.NewV4()).String(),
		Status:    entities.StatusCreated.String(),
		Job:       FakeJobModel(0),
		CreatedAt: time.Now(),
	}
}

func FakeAccountModel() *models.Account {
	return &models.Account{
		Alias:               utils.RandString(10),
		TenantID:            "tenantID",
		Address:             ethcommon.HexToAddress(utils.RandHexString(12)).String(),
		PublicKey:           ethcommon.HexToHash(utils.RandHexString(12)).String(),
		CompressedPublicKey: ethcommon.HexToHash(utils.RandHexString(12)).String(),
		Attributes: map[string]string{
			"attr1": "val1",
			"attr2": "val2",
		},
	}
}

func FakeFaucetModel() *models.Faucet {
	return &models.Faucet{
		UUID:            uuid.Must(uuid.NewV4()).String(),
		TenantID:        "tenantID",
		Name:            "faucet-mainnet",
		ChainRule:       uuid.Must(uuid.NewV4()).String(),
		CreditorAccount: "0x5Cc634233E4a454d47aACd9fC68801482Fb02610",
		MaxBalance:      "60000000000000000",
		Amount:          "100000000000000000",
		Cooldown:        "10s",
	}
}

func FakeChainModel() *models.Chain {
	return &models.Chain{
		UUID:                      uuid.Must(uuid.NewV4()).String(),
		Name:                      "chain" + utils.RandString(5),
		TenantID:                  "tenantID",
		URLs:                      []string{"http://ganache:8545"},
		ChainID:                   "888",
		ListenerDepth:             0,
		ListenerCurrentBlock:      1,
		ListenerStartingBlock:     0,
		ListenerBackOffDuration:   "5s",
		ListenerExternalTxEnabled: false,
		Labels: map[string]string{
			"label1": "value1",
		},
	}
}
