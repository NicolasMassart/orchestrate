package formatters

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/consensys/orchestrate/pkg/errors"
	infra "github.com/consensys/orchestrate/src/infra/api"
	log "github.com/sirupsen/logrus"

	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
)

func FormatChainResponse(chain *entities.Chain) *types.ChainResponse {
	res := &types.ChainResponse{
		UUID:                      chain.UUID,
		Name:                      chain.Name,
		TenantID:                  chain.TenantID,
		OwnerID:                   chain.OwnerID,
		URLs:                      chain.URLs,
		ChainID:                   chain.ChainID.String(),
		ListenerDepth:             chain.ListenerDepth,
		ListenerCurrentBlock:      chain.ListenerCurrentBlock,
		ListenerStartingBlock:     chain.ListenerStartingBlock,
		ListenerBackOffDuration:   chain.ListenerBackOffDuration.String(),
		ListenerExternalTxEnabled: chain.ListenerExternalTxEnabled,
		Labels:                    chain.Labels,
		CreatedAt:                 chain.CreatedAt,
		UpdatedAt:                 chain.UpdatedAt,
	}

	if chain.PrivateTxManager != nil {
		res.PrivateTxManager = FormatPrivateTxManagerResponse(chain.PrivateTxManager)
	}

	return res
}

func FormatPrivateTxManagerResponse(txManager *entities.PrivateTxManager) *types.PrivateTxManagerResponse {
	return &types.PrivateTxManagerResponse{
		URL:       txManager.URL,
		Type:      txManager.Type,
		CreatedAt: txManager.CreatedAt,
	}
}

func FormatRegisterChainRequest(request *types.RegisterChainRequest, fromLatest bool) (*entities.Chain, error) {
	chain := &entities.Chain{
		Name:                      request.Name,
		URLs:                      request.URLs,
		ListenerDepth:             request.Listener.Depth,
		ListenerExternalTxEnabled: request.Listener.ExternalTxEnabled,
		Labels:                    request.Labels,
	}

	if request.Listener.BackOffDuration == "" {
		chain.ListenerBackOffDuration, _ = time.ParseDuration("5s")
	} else {
		var err error
		chain.ListenerBackOffDuration, err = time.ParseDuration(request.Listener.BackOffDuration)
		if err != nil {
			return nil, err
		}
	}

	if !fromLatest {
		startingBlock, err := strconv.ParseUint(request.Listener.FromBlock, 10, 64)
		if err != nil {
			errMessage := "fromBlock must be an integer value"
			log.WithField("from_block", fmt.Sprintf("%q", request.Listener.FromBlock)).Error(errMessage)
			return nil, errors.InvalidFormatError(errMessage)
		}
		chain.ListenerStartingBlock = startingBlock
		chain.ListenerCurrentBlock = startingBlock
	}

	if request.PrivateTxManager != nil {
		chain.PrivateTxManager = &entities.PrivateTxManager{
			URL:  request.PrivateTxManager.URL,
			Type: request.PrivateTxManager.Type,
		}
	}

	return chain, nil
}

func FormatUpdateChainRequest(request *types.UpdateChainRequest, uuid string) (*entities.Chain, error) {
	chain := &entities.Chain{
		UUID:   uuid,
		Name:   request.Name,
		Labels: request.Labels,
	}

	if request.Listener != nil {
		if request.Listener.BackOffDuration != "" {
			var err error
			chain.ListenerBackOffDuration, err = time.ParseDuration(request.Listener.BackOffDuration)
			if err != nil {
				return nil, err
			}
		}
		chain.ListenerDepth = request.Listener.Depth
		chain.ListenerExternalTxEnabled = request.Listener.ExternalTxEnabled
		chain.ListenerCurrentBlock = request.Listener.CurrentBlock
	}

	if request.PrivateTxManager != nil {
		chain.PrivateTxManager = &entities.PrivateTxManager{
			URL:  request.PrivateTxManager.URL,
			Type: request.PrivateTxManager.Type,
		}
	}

	return chain, nil
}

func FormatChainFiltersRequest(req *http.Request) (*entities.ChainFilters, error) {
	filters := &entities.ChainFilters{}

	qNames := req.URL.Query().Get("names")
	if qNames != "" {
		filters.Names = strings.Split(qNames, ",")
	}

	if err := infra.GetValidator().Struct(filters); err != nil {
		return nil, err
	}

	return filters, nil
}

func ChainResponseToEntity(chain *types.ChainResponse) *entities.Chain {
	// Cannot fail as the duration coming from a response is expected to be valid
	chainID, _ := new(big.Int).SetString(chain.ChainID, 10)
	listenerBackOffDuration, _ := time.ParseDuration(chain.ListenerBackOffDuration)
	return &entities.Chain{
		UUID:                      chain.UUID,
		Name:                      chain.Name,
		TenantID:                  chain.TenantID,
		OwnerID:                   chain.OwnerID,
		URLs:                      chain.URLs,
		ChainID:                   chainID,
		ListenerDepth:             chain.ListenerDepth,
		ListenerCurrentBlock:      chain.ListenerCurrentBlock,
		ListenerStartingBlock:     chain.ListenerStartingBlock,
		ListenerBackOffDuration:   listenerBackOffDuration,
		ListenerExternalTxEnabled: chain.ListenerExternalTxEnabled,
		PrivateTxManager:          PrivateTxManagerResponseToEntity(chain.PrivateTxManager),
		Labels:                    chain.Labels,
		CreatedAt:                 chain.CreatedAt,
		UpdatedAt:                 chain.UpdatedAt,
	}
}

func PrivateTxManagerResponseToEntity(privTxMngr *types.PrivateTxManagerResponse) *entities.PrivateTxManager {
	if privTxMngr == nil {
		return nil
	}
	// Cannot fail as the duration coming from a response is expected to be valid
	return &entities.PrivateTxManager{
		URL:       privTxMngr.URL,
		CreatedAt: privTxMngr.CreatedAt,
		Type:      privTxMngr.Type,
	}
}
