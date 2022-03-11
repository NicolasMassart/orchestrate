package formatters

import (
	"net/http"
	"strings"

	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"
)

func FormatRegisterFaucetRequest(request *types.RegisterFaucetRequest) *entities.Faucet {
	return &entities.Faucet{
		Name:            request.Name,
		ChainRule:       request.ChainRule,
		CreditorAccount: request.CreditorAccount,
		MaxBalance:      request.MaxBalance,
		Amount:          request.Amount,
		Cooldown:        request.Cooldown,
	}
}

func FormatUpdateFaucetRequest(request *types.UpdateFaucetRequest, uuid string) *entities.Faucet {
	faucet := &entities.Faucet{
		UUID:       uuid,
		Name:       request.Name,
		ChainRule:  request.ChainRule,
		MaxBalance: request.MaxBalance,
		Amount:     request.Amount,
		Cooldown:   request.Cooldown,
	}

	if request.CreditorAccount != nil {
		faucet.CreditorAccount = *request.CreditorAccount
	}

	return faucet
}

func FormatFaucetResponse(faucet *entities.Faucet) *types.FaucetResponse {
	return &types.FaucetResponse{
		UUID:            faucet.UUID,
		Name:            faucet.Name,
		TenantID:        faucet.TenantID,
		ChainRule:       faucet.ChainRule,
		CreditorAccount: faucet.CreditorAccount.String(),
		MaxBalance:      faucet.MaxBalance.String(),
		Amount:          faucet.Amount.String(),
		Cooldown:        faucet.Cooldown,
		CreatedAt:       faucet.CreatedAt,
		UpdatedAt:       faucet.UpdatedAt,
	}
}

func FormatFaucetFilters(req *http.Request) (*entities.FaucetFilters, error) {
	filters := &entities.FaucetFilters{}

	qNames := req.URL.Query().Get("names")
	if qNames != "" {
		filters.Names = strings.Split(qNames, ",")
	}

	qChainRule := req.URL.Query().Get("chain_rule")
	if qChainRule != "" {
		filters.ChainRule = qChainRule
	}

	if err := infra.GetValidator().Struct(filters); err != nil {
		return nil, err
	}

	return filters, nil
}
