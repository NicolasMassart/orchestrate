package controllers

import (
	"encoding/json"
	"net/http"

	infra "github.com/consensys/orchestrate/src/infra/api"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/gorilla/mux"
)

// Hack for swagger generation
type ChainsController struct {
	ucs usecases.ChainUseCases
}

func NewChainsController(chainUCs usecases.ChainUseCases) *ChainsController {
	return &ChainsController{ucs: chainUCs}
}

func (c *ChainsController) Append(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/chains").HandlerFunc(c.search)
	router.Methods(http.MethodGet).Path("/chains/{uuid}").HandlerFunc(c.getOne)
	router.Methods(http.MethodPost).Path("/chains").HandlerFunc(c.register)
	router.Methods(http.MethodPatch).Path("/chains/{uuid}").HandlerFunc(c.update)
	router.Methods(http.MethodDelete).Path("/chains/{uuid}").HandlerFunc(c.delete)
}

// @Summary   Retrieves a list of all registered chains
// @Tags      Chains
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Success   200  {array}   api.ChainResponse{privateTxManager=entities.PrivateTxManager}
// @Failure   400  {object}  infra.ErrorResponse  "Invalid request"
// @Failure   500  {object}  infra.ErrorResponse  "Internal server error"
// @Router    /chains [get]
func (c *ChainsController) search(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	filters, err := formatters.FormatChainFiltersRequest(request)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	chains, err := c.ucs.Search().Execute(ctx, filters, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	response := []*api.ChainResponse{}
	for _, chain := range chains {
		response = append(response, formatters.FormatChainResponse(chain))
	}

	_ = json.NewEncoder(rw).Encode(response)
}

// @Summary   Retrieves a chain by ID
// @Tags      Chains
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid  path      string  true  "ID of the chain"
// @Success   200   {object}  api.ChainResponse{}
// @Failure   400   {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404   {object}  infra.ErrorResponse  "Chain not found"
// @Failure   500   {object}  infra.ErrorResponse  "Internal server error"
// @Router    /chains/{uuid} [get]
func (c *ChainsController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	chain, err := c.ucs.Get().Execute(ctx, mux.Vars(request)["uuid"], multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatChainResponse(chain))
}

// @Summary   Updates a chain by ID
// @Tags      Chains
// @Accept    json
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid     path      string                                                                                                   true  "ID of the chain"
// @Param     request  body      api.UpdateChainRequest{listener=api.UpdateListenerRequest}  true  "Chain update request"
// @Success   200      {object}  api.ChainResponse{}
// @Failure   400      {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404      {object}  infra.ErrorResponse  "Chain not found"
// @Failure   500      {object}  infra.ErrorResponse  "Internal server error"
// @Router    /chains/{uuid} [patch]
func (c *ChainsController) update(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	chainRequest := &api.UpdateChainRequest{}
	err := infra.UnmarshalBody(request.Body, chainRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	uuid := mux.Vars(request)["uuid"]
	nextChain, err := formatters.FormatUpdateChainRequest(chainRequest, uuid)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}
	chain, err := c.ucs.Update().Execute(ctx, nextChain, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatChainResponse(chain))
}

// @Summary   Registers a new chain
// @Tags      Chains
// @Accept    json
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     request  body      api.RegisterChainRequest{listener=api.RegisterListenerRequest}  true  "Chain registration request."
// @Success   200      {object}  api.ChainResponse{privateTxManager=entities.PrivateTxManager}
// @Failure   400      {object}  infra.ErrorResponse  "Invalid request"
// @Failure   500      {object}  infra.ErrorResponse  "Internal server error"
// @Router    /chains [post]
func (c *ChainsController) register(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	chainRequest := &api.RegisterChainRequest{}
	err := infra.UnmarshalBody(request.Body, chainRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	fromLatest := chainRequest.Listener.FromBlock == "" || chainRequest.Listener.FromBlock == "latest"
	chain, err := formatters.FormatRegisterChainRequest(chainRequest, fromLatest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	chain, err = c.ucs.Register().Execute(ctx, chain, fromLatest, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatChainResponse(chain))
}

// @Summary   Deletes a chain by ID
// @Tags      Chains
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid  path  string  true  "ID of the chain"
// @Success   204
// @Failure   400  {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404  {object}  infra.ErrorResponse  "Chain not found"
// @Failure   500  {object}  infra.ErrorResponse  "Internal server error"
// @Router    /chains/{uuid} [delete]
func (c *ChainsController) delete(rw http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	uuid := mux.Vars(request)["uuid"]
	err := c.ucs.Delete().Execute(ctx, uuid, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
