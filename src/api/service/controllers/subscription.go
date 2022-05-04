package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

type SubscriptionsController struct {
	ucs usecases.SubscriptionUseCases
}

func NewSubscriptionsController(subscriptionUCs usecases.SubscriptionUseCases) *SubscriptionsController {
	return &SubscriptionsController{ucs: subscriptionUCs}
}

// Append Add routes to router
func (c *SubscriptionsController) Append(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/subscriptions").HandlerFunc(c.search)
	router.Methods(http.MethodPost).Path("/subscriptions/{address}").HandlerFunc(c.create)
	router.Methods(http.MethodGet).Path("/subscriptions/{uuid}").HandlerFunc(c.getOne)
	router.Methods(http.MethodPatch).Path("/subscriptions/{uuid}").HandlerFunc(c.update)
	router.Methods(http.MethodDelete).Path("/subscriptions/{uuid}").HandlerFunc(c.delete)
}

// @Summary      Creates a new Subscription Event stream
// @Description  Creates a new Subscription Event stream
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.CreateSubscriptionRequest  true  "Subscription creation request"
// @Success      200      {object}  api.SubscriptionResponse       "Subscription object"
// @Failure      400      {object}  infra.ErrorResponse    "Invalid request"
// @Failure      401      {object}  infra.ErrorResponse    "Unauthorized"
// @Failure      500      {object}  infra.ErrorResponse    "Internal server error"
// @Router       /subscriptions [post]
func (c *SubscriptionsController) create(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	req := &api.CreateSubscriptionRequest{}
	err := infra.UnmarshalBody(request.Body, req)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	qAddress := request.URL.Query().Get("address")
	es, err := c.ucs.Create().Execute(ctx, req.ToEntity(ethcommon.HexToAddress(qAddress)), req.Chain, req.EventStream, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewSubscriptionResponse(es))
}

// @Summary      Fetch an event stream by uuid
// @Description  Fetch a single event stream by uuid
// @Tags         Subscriptions
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        address  path      string                  true  "event stream uuid"
// @Success      200      {object}  api.SubscriptionResponse     "Event stream found"
// @Failure      404      {object}  infra.ErrorResponse  "Event stream not found"
// @Failure      401      {object}  infra.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  infra.ErrorResponse  "Internal server error"
// @Router       /subscriptions/{uuid} [get]
func (c *SubscriptionsController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	sub, err := c.ucs.Get().Execute(ctx, mux.Vars(request)["uuid"], multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewSubscriptionResponse(sub))
}

// @Summary      Search subscription by provided filters
// @Description  Get a list of filtered subscriptions
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        names   query     []string false "List of subscription addresses"  collectionFormat(csv)
// @Param        chain_uuid  query     string false  "Chain UUID"
// @Param        tenant_id  query     string false  "Tenant ID"
// @Success      200      {array}   api.SubscriptionResponse "List of subscription found"
// @Failure      400      {object}  infra.ErrorResponse  "Invalid filter in the request"
// @Failure      401      {object}  infra.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  infra.ErrorResponse  "Internal server error"
// @Router       /subscriptions [get]
func (c *SubscriptionsController) search(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	filters := &entities.SubscriptionFilters{}

	qAddresses := request.URL.Query().Get("address")
	if qAddresses != "" {
		for _, addr := range strings.Split(qAddresses, ",") {
			filters.Addresses = append(filters.Addresses, ethcommon.HexToAddress(addr))
		}
	}

	qChainUUID := request.URL.Query().Get("chain_uuid")
	if qChainUUID != "" {
		filters.ChainUUID = qChainUUID
	}

	qTenantID := request.URL.Query().Get("tenant_id")
	if qTenantID != "" {
		filters.TenantID = qTenantID
	}

	if err := infra.GetValidator().Struct(filters); err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	subscriptions, err := c.ucs.Search().Execute(ctx, filters, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewSubscriptionResponses(subscriptions))
}

// @Summary      Update subscription by uuid
// @Description  Update a subscription by uuid
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.request  true  "Kafka Event stream update request"
// @Param        uuid  path      string                    true  "event stream uuid"
// @Success      200      {object}  api.SubscriptionResponse       "Subscription found"
// @Failure      400      {object}  infra.ErrorResponse    "Invalid request"
// @Failure      401      {object}  infra.ErrorResponse    "Unauthorized"
// @Failure      404      {object}  infra.ErrorResponse    "Account not found"
// @Failure      500      {object}  infra.ErrorResponse    "Internal server error"
// @Router       /subscriptions/{uuid} [patch]
func (c *SubscriptionsController) update(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	req := &api.UpdateSubscriptionRequest{}
	err := infra.UnmarshalBody(request.Body, req)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	sub, err := c.ucs.Update().Execute(ctx, req.ToEntity(mux.Vars(request)["uuid"]), req.EventStream, multitenancy.UserInfoValue(ctx))

	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewSubscriptionResponse(sub))
}

// @Summary   Deletes a subscription by uuid
// @Tags      Subscriptions
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid  path  string  true  "uuid of the subscription"
// @Success   204
// @Failure   400  {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404  {object}  infra.ErrorResponse  "Subscription not found"
// @Failure   500  {object}  infra.ErrorResponse  "Internal server error"
// @Router    /subscriptions/{uuid} [delete]
func (c *SubscriptionsController) delete(rw http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	err := c.ucs.Delete().Execute(ctx, mux.Vars(request)["uuid"], multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
