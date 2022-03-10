package parsers

import (
	"math/big"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/store/models"
	"github.com/consensys/orchestrate/src/entities"
)

func NewChainEntity(chainModel *models.Chain) *entities.Chain {
	listenerBackOffDuration, _ := time.ParseDuration(chainModel.ListenerBackOffDuration)

	chain := &entities.Chain{
		UUID:                      chainModel.UUID,
		Name:                      chainModel.Name,
		TenantID:                  chainModel.TenantID,
		OwnerID:                   chainModel.OwnerID,
		URLs:                      chainModel.URLs,
		PrivateTxManagerURL:       chainModel.PrivateTxManagerURL,
		ChainID:                   (*big.Int)(utils.StringBigIntToHex(chainModel.ChainID)),
		ListenerDepth:             chainModel.ListenerDepth,
		ListenerCurrentBlock:      chainModel.ListenerCurrentBlock,
		ListenerStartingBlock:     chainModel.ListenerStartingBlock,
		ListenerBackOffDuration:   listenerBackOffDuration,
		ListenerExternalTxEnabled: chainModel.ListenerExternalTxEnabled,
		Labels:                    chainModel.Labels,
		CreatedAt:                 chainModel.CreatedAt,
		UpdatedAt:                 chainModel.UpdatedAt,
	}

	return chain
}

func NewChainEntityArr(chainModels []*models.Chain) []*entities.Chain {
	res := []*entities.Chain{}
	for _, model := range chainModels {
		res = append(res, NewChainEntity(model))
	}

	return res
}

func NewChainModel(chain *entities.Chain) *models.Chain {
	chainModel := &models.Chain{
		UUID:                      chain.UUID,
		Name:                      chain.Name,
		TenantID:                  chain.TenantID,
		OwnerID:                   chain.OwnerID,
		URLs:                      chain.URLs,
		PrivateTxManagerURL:       chain.PrivateTxManagerURL,
		ListenerDepth:             chain.ListenerDepth,
		ListenerCurrentBlock:      chain.ListenerCurrentBlock,
		ListenerStartingBlock:     chain.ListenerStartingBlock,
		ListenerExternalTxEnabled: chain.ListenerExternalTxEnabled,
		Labels:                    chain.Labels,
		CreatedAt:                 chain.CreatedAt,
		UpdatedAt:                 chain.UpdatedAt,
	}

	if chain.ListenerBackOffDuration.Milliseconds() > 0 {
		chainModel.ListenerBackOffDuration = chain.ListenerBackOffDuration.String()
	}

	if chain.ChainID != nil {
		chainModel.ChainID = chain.ChainID.String()
	}

	return chainModel
}
