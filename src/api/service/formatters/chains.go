package formatters

import (
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"
)

func FormatChainResponse(chain *entities.Chain) *types.ChainResponse {
	res := &types.ChainResponse{
		UUID:                      chain.UUID,
		Name:                      chain.Name,
		TenantID:                  chain.TenantID,
		OwnerID:                   chain.OwnerID,
		URLs:                      chain.URLs,
		PrivateTxManagerURL:       chain.PrivateTxManagerURL,
		ChainID:                   chain.ChainID.Uint64(),
		ListenerDepth:             chain.ListenerDepth,
		ListenerBlockTimeDuration: chain.ListenerBlockTimeDuration.String(),
		Labels:                    chain.Labels,
		CreatedAt:                 chain.CreatedAt,
		UpdatedAt:                 chain.UpdatedAt,
	}

	return res
}

func FormatRegisterChainRequest(request *types.RegisterChainRequest) (*entities.Chain, error) {
	chain := &entities.Chain{
		Name:                request.Name,
		URLs:                request.URLs,
		PrivateTxManagerURL: request.PrivateTxManagerURL,
		ListenerDepth:       request.Listener.Depth,
		Labels:              request.Labels,
	}

	if request.Listener.BlockTimeDuration == "" {
		chain.ListenerBlockTimeDuration, _ = time.ParseDuration("5s")
	} else {
		var err error
		chain.ListenerBlockTimeDuration, err = time.ParseDuration(request.Listener.BlockTimeDuration)
		if err != nil {
			return nil, err
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
		if request.Listener.BlockTimeDuration != "" {
			var err error
			chain.ListenerBlockTimeDuration, err = time.ParseDuration(request.Listener.BlockTimeDuration)
			if err != nil {
				return nil, err
			}
		}
		chain.ListenerDepth = request.Listener.Depth
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
	listenerBackOffDuration, _ := time.ParseDuration(chain.ListenerBlockTimeDuration)
	return &entities.Chain{
		UUID:                      chain.UUID,
		Name:                      chain.Name,
		TenantID:                  chain.TenantID,
		OwnerID:                   chain.OwnerID,
		URLs:                      chain.URLs,
		ChainID:                   new(big.Int).SetUint64(chain.ChainID),
		ListenerDepth:             chain.ListenerDepth,
		ListenerBlockTimeDuration: listenerBackOffDuration,
		Labels:                    chain.Labels,
		CreatedAt:                 chain.CreatedAt,
		UpdatedAt:                 chain.UpdatedAt,
	}
}
