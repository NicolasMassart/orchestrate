package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	infra "github.com/consensys/orchestrate/src/infra/api"

	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/gorilla/mux"
)

type FaucetsController struct {
	ucs usecases.FaucetUseCases
}

func NewFaucetsController(ucs usecases.FaucetUseCases) *FaucetsController {
	return &FaucetsController{ucs: ucs}
}

func (c *FaucetsController) Append(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/faucets").HandlerFunc(c.search)
	router.Methods(http.MethodGet).Path("/faucets/{uuid}").HandlerFunc(c.getOne)
	router.Methods(http.MethodPost).Path("/faucets").HandlerFunc(c.register)
	router.Methods(http.MethodPatch).Path("/faucets/{uuid}").HandlerFunc(c.update)
	router.Methods(http.MethodDelete).Path("/faucets/{uuid}").HandlerFunc(c.delete)
}

// @Summary   Retrieves a list of all registered faucets
// @Tags      Faucets
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Success   200  {array}   api.FaucetResponse
// @Failure   400  {object}  infra.ErrorResponse  "Invalid request"
// @Failure   500  {object}  infra.ErrorResponse  "Internal server error"
// @Router    /faucets [get]
func (c *FaucetsController) search(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	filters, err := formatters.FormatFaucetFilters(request)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	faucets, err := c.ucs.Search().Execute(ctx, filters, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	response := []*api.FaucetResponse{}
	for _, faucet := range faucets {
		response = append(response, formatters.FormatFaucetResponse(faucet))
	}

	_ = json.NewEncoder(rw).Encode(response)
}

// @Summary  Retrieves a faucet by ID
// @Tags     Faucets
// @Produce  json
// @Param    uuid  path      string  true  "ID of the faucet"
// @Success  200   {object}  api.FaucetResponse
// @Failure  404   {object}  infra.ErrorResponse  "Faucet not found"
// @Failure  500   {object}  infra.ErrorResponse  "Internal server error"
// @Router   /faucets/{uuid} [get]
func (c *FaucetsController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	faucet, err := c.ucs.Get().Execute(ctx, mux.Vars(request)["uuid"], multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatFaucetResponse(faucet))
}

// @Summary   Registers a new faucet
// @Tags      Faucets
// @Accept    json
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     request  body      api.RegisterFaucetRequest  true  "Faucet registration request"
// @Success   200      {object}  api.FaucetResponse
// @Failure   400      {object}  infra.ErrorResponse  "Invalid request"
// @Failure   422      {object}  infra.ErrorResponse  "Unprocessable entity"
// @Failure   500      {object}  infra.ErrorResponse  "Internal server error"
// @Router    /faucets [post]
func (c *FaucetsController) register(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	faucetRequest := &api.RegisterFaucetRequest{}
	err := infra.UnmarshalBody(request.Body, faucetRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	faucet, err := c.ucs.Register().Execute(ctx, formatters.FormatRegisterFaucetRequest(faucetRequest), multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatFaucetResponse(faucet))
}

// @Summary   Updates a faucet by ID
// @Tags      Faucets
// @Accept    json
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid     path      string                   true  "ID of the faucet"
// @Param     request  body      api.UpdateFaucetRequest  true  "Faucet update request"
// @Success   200      {object}  api.FaucetResponse
// @Failure   400      {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404      {object}  infra.ErrorResponse  "Faucet not found"
// @Failure   422      {object}  infra.ErrorResponse  "Unprocessable entity"
// @Failure   500      {object}  infra.ErrorResponse  "Internal server error"
// @Router    /faucets/{uuid} [patch]
func (c *FaucetsController) update(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	faucetRequest := &api.UpdateFaucetRequest{}
	err := infra.UnmarshalBody(request.Body, faucetRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	uuid := mux.Vars(request)["uuid"]
	faucet, err := c.ucs.Update().Execute(ctx, formatters.FormatUpdateFaucetRequest(faucetRequest, uuid),
		multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatFaucetResponse(faucet))
}

// @Summary   Deletes a faucet by ID
// @Tags      Faucets
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid  path  string  true  "ID of the faucet"
// @Success   204
// @Failure   400  {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404  {object}  infra.ErrorResponse  "Faucet not found"
// @Failure   500  {object}  infra.ErrorResponse  "Internal server error"
// @Router    /faucets/{uuid} [delete]
func (c *FaucetsController) delete(rw http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	uuid := mux.Vars(request)["uuid"]
	err := c.ucs.Delete().Execute(ctx, uuid, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
