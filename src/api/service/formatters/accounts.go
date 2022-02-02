package formatters

import (
	"net/http"
	"strings"

	api "github.com/consensys/orchestrate/src/api/service/types"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
)

func FormatCreateAccountRequest(req *api.CreateAccountRequest, defaultStoreID string) *entities.Account {
	acc := &entities.Account{
		Alias:      req.Alias,
		Attributes: req.Attributes,
		StoreID:    req.StoreID,
	}

	if acc.StoreID == "" {
		acc.StoreID = defaultStoreID
	}

	return acc
}

func FormatImportAccountRequest(req *api.ImportAccountRequest, defaultStoreID string) *entities.Account {
	acc := &entities.Account{
		Alias:      req.Alias,
		Attributes: req.Attributes,
		StoreID:    req.StoreID,
	}

	if acc.StoreID == "" {
		acc.StoreID = defaultStoreID
	}

	return acc
}

func FormatUpdateAccountRequest(req *api.UpdateAccountRequest) *entities.Account {
	return &entities.Account{
		Alias:      req.Alias,
		Attributes: req.Attributes,
		StoreID:    req.StoreID,
	}
}

func FormatAccountResponse(iden *entities.Account) *api.AccountResponse {
	return &api.AccountResponse{
		Alias:               iden.Alias,
		Attributes:          iden.Attributes,
		Address:             iden.Address.String(),
		PublicKey:           iden.PublicKey.String(),
		CompressedPublicKey: iden.CompressedPublicKey.String(),
		TenantID:            iden.TenantID,
		OwnerID:             iden.OwnerID,
		StoreID:             iden.StoreID,
		CreatedAt:           iden.CreatedAt,
		UpdatedAt:           iden.UpdatedAt,
	}
}

func FormatAccountFilterRequest(req *http.Request) (*entities.AccountFilters, error) {
	filters := &entities.AccountFilters{}

	qAliases := req.URL.Query().Get("aliases")
	if qAliases != "" {
		filters.Aliases = strings.Split(qAliases, ",")
	}

	if err := utils.GetValidator().Struct(filters); err != nil {
		return nil, err
	}

	return filters, nil
}
