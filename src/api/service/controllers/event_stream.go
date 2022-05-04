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
	"github.com/gorilla/mux"
)

type EventStreamsController struct {
	ucs usecases.EventStreamsUseCases
}

func NewEventStreamsController(eventStreamUCs usecases.EventStreamsUseCases) *EventStreamsController {
	return &EventStreamsController{ucs: eventStreamUCs}
}

// Append Add routes to router
func (c *EventStreamsController) Append(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/eventstreams").HandlerFunc(c.search)
	router.Methods(http.MethodPost).Path("/eventstreams").HandlerFunc(c.create)
	router.Methods(http.MethodGet).Path("/eventstreams/{uuid}").HandlerFunc(c.getOne)
	router.Methods(http.MethodPatch).Path("/eventstreams/{uuid}").HandlerFunc(c.update)
	router.Methods(http.MethodDelete).Path("/eventstreams/{uuid}").HandlerFunc(c.delete)
}

// @Summary      Creates a new Event stream
// @Description  Creates a new Event stream
// @Tags         Event Streams
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.CreateEventStreamRequest  true  "Event stream creation request"
// @Success      200      {object}  api.EventStreamResponse       "Event stream object"
// @Failure      400      {object}  infra.ErrorResponse    "Invalid request"
// @Failure      401      {object}  infra.ErrorResponse    "Unauthorized"
// @Failure      500      {object}  infra.ErrorResponse    "Internal server error"
// @Router       /eventstreams [post]
func (c *EventStreamsController) create(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	req := &api.CreateEventStreamRequest{}
	err := infra.UnmarshalBody(request.Body, req)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if !req.Validate() {
		infra.WriteError(rw, "invalid event stream creation request", http.StatusBadRequest)
		return
	}

	es, err := c.ucs.Create().Execute(ctx, req.ToEntity(), req.Chain, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewEventStreamResponse(es))
}

// @Summary      Fetch an event stream by uuid
// @Description  Fetch a single event stream by uuid
// @Tags         Event Streams
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        address  path      string                  true  "event stream uuid"
// @Success      200      {object}  api.EventStreamResponse     "Event stream found"
// @Failure      404      {object}  infra.ErrorResponse  "Event stream not found"
// @Failure      401      {object}  infra.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  infra.ErrorResponse  "Internal server error"
// @Router       /eventstreams/{uuid} [get]
func (c *EventStreamsController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	es, err := c.ucs.Get().Execute(ctx, mux.Vars(request)["uuid"], multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewEventStreamResponse(es))
}

// @Summary      Search event streams by provided filters
// @Description  Get a list of filtered event streams
// @Tags         Event Streams
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        names   query     []string false "List of event stream names"  collectionFormat(csv)
// @Param        chain_uuid  query     string false  "Chain UUID"
// @Param        tenant_id  query     string false  "Tenant ID"
// @Success      200      {array}   api.EventStreamResponse "List of event streams found"
// @Failure      400      {object}  infra.ErrorResponse  "Invalid filter in the request"
// @Failure      401      {object}  infra.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  infra.ErrorResponse  "Internal server error"
// @Router       /eventstreams [get]
func (c *EventStreamsController) search(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	filters := &entities.EventStreamFilters{}

	qTxHashes := request.URL.Query().Get("names")
	if qTxHashes != "" {
		filters.Names = strings.Split(qTxHashes, ",")
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

	eventStreams, err := c.ucs.Search().Execute(ctx, filters, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewEventStreamResponses(eventStreams))
}

// @Summary      Update event stream by uuid
// @Description  Update a specific event stream by uuid
// @Tags         Event Streams
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.UpdateEventStreamRequest  true  "Event stream update request"
// @Param        uuid  path      string                    true  "event stream uuid"
// @Success      200      {object}  api.EventStreamResponse       "Event stream found"
// @Failure      400      {object}  infra.ErrorResponse    "Invalid request"
// @Failure      401      {object}  infra.ErrorResponse    "Unauthorized"
// @Failure      404      {object}  infra.ErrorResponse    "Account not found"
// @Failure      500      {object}  infra.ErrorResponse    "Internal server error"
// @Router       /eventstreams/{uuid} [patch]
func (c *EventStreamsController) update(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	req := &api.UpdateEventStreamRequest{}
	err := infra.UnmarshalBody(request.Body, req)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	es, err := c.ucs.Update().Execute(ctx, req.ToEntity(mux.Vars(request)["uuid"]), multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(api.NewEventStreamResponse(es))
}

// @Summary   Deletes an event stream by uuid
// @Tags      Event Streams
// @Produce   json
// @Security  ApiKeyAuth
// @Security  JWTAuth
// @Param     uuid  path  string  true  "uuid of the event stream"
// @Success   204
// @Failure   400  {object}  infra.ErrorResponse  "Invalid request"
// @Failure   404  {object}  infra.ErrorResponse  "Event Stream not found"
// @Failure   500  {object}  infra.ErrorResponse  "Internal server error"
// @Router    /eventstreams/{uuid} [delete]
func (c *EventStreamsController) delete(rw http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	err := c.ucs.Delete().Execute(ctx, mux.Vars(request)["uuid"], multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
