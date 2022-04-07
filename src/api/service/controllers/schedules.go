package controllers

import (
	"encoding/json"
	"net/http"

	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/gorilla/mux"
)

type SchedulesController struct {
	ucs usecases.ScheduleUseCases
}

func NewSchedulesController(useCases usecases.ScheduleUseCases) *SchedulesController {
	return &SchedulesController{
		ucs: useCases,
	}
}

func (c *SchedulesController) Append(router *mux.Router) {
	router.Methods(http.MethodPost).Path("/schedules").HandlerFunc(c.create)
	router.Methods(http.MethodGet).Path("/schedules/{uuid}").HandlerFunc(c.getOne)
	router.Methods(http.MethodGet).Path("/schedules").HandlerFunc(c.getAll)
}

// @Summary      Creates a new Schedule
// @Description  Creates a new schedule
// @Tags         Schedules
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.CreateScheduleRequest                     true  "Schedule creation request"
// @Success      200      {object}  api.ScheduleResponse{jobs=[]api.JobResponse}  "Created schedule"
// @Failure      400      {object}  httputil.ErrorResponse                        "Invalid request"
// @Failure      422      {object}  httputil.ErrorResponse                        "Unprocessable parameters were sent"
// @Failure      500      {object}  httputil.ErrorResponse                        "Internal server error"
// @Router       /schedules [post]
func (c *SchedulesController) create(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	scheduleRequest := &api.CreateScheduleRequest{}
	err := infra.UnmarshalBody(request.Body, scheduleRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	scheduleEntity, err := c.ucs.CreateSchedule().Execute(ctx, &entities.Schedule{}, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	response := formatters.FormatScheduleResponse(scheduleEntity)
	_ = json.NewEncoder(rw).Encode(response)
}

// @Summary      Fetch a schedule by uuid
// @Description  Fetch a single schedule by uuid
// @Tags         Schedules
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        uuid  path      string                  true  "UUID of the schedule"
// @Success      200   {object}  api.ScheduleResponse    "Schedule found"
// @Failure      404   {object}  httputil.ErrorResponse  "Schedule not found"
// @Failure      500   {object}  httputil.ErrorResponse  "Internal server error"
// @Router       /schedules/{uuid} [get]
func (c *SchedulesController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	uuid := mux.Vars(request)["uuid"]

	scheduleEntity, err := c.ucs.GetSchedule().Execute(ctx, uuid, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	response := formatters.FormatScheduleResponse(scheduleEntity)
	_ = json.NewEncoder(rw).Encode(response)
}

// @Summary      Get all schedules
// @Description  Get all schedules
// @Tags         Schedules
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Success      200  {array}   api.ScheduleResponse    "List of schedules found"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Router       /schedules [get]
func (c *SchedulesController) getAll(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	schedules, err := c.ucs.SearchSchedules().Execute(ctx, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	response := []*api.ScheduleResponse{}
	for _, schedule := range schedules {
		response = append(response, formatters.FormatScheduleResponse(schedule))
	}

	_ = json.NewEncoder(rw).Encode(response)
}
