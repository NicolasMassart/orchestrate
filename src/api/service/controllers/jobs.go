package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/consensys/orchestrate/src/entities"
	infra "github.com/consensys/orchestrate/src/infra/api"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/gorilla/mux"
)

var _ entities.ETHTransaction

type JobsController struct {
	ucs usecases.JobUseCases
}

func NewJobsController(useCases usecases.JobUseCases) *JobsController {
	return &JobsController{
		ucs: useCases,
	}
}

func (c *JobsController) Append(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/jobs").HandlerFunc(c.search)
	router.Methods(http.MethodPost).Path("/jobs").HandlerFunc(c.create)
	router.Methods(http.MethodGet).Path("/jobs/{uuid}").HandlerFunc(c.getOne)
	router.Methods(http.MethodPatch).Path("/jobs/{uuid}").HandlerFunc(c.update)
	router.Methods(http.MethodPut).Path("/jobs/{uuid}/start").HandlerFunc(c.start)
	router.Methods(http.MethodPut).Path("/jobs/{uuid}/resend").HandlerFunc(c.resend)
}

// @Summary      Search jobs by provided filters
// @Description  Get a list of filtered jobs
// @Tags         Jobs
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        tx_hashes   query     []string                                                                                                                                                              false  "List of transaction hashes"  collectionFormat(csv)
// @Param        chain_uuid  query     string                                                                                                                                                                false  "Chain UUID"
// @Success      200         {array}   api.JobResponse  "List of Jobs found"
// @Failure      400         {object}  infra.ErrorResponse                                                                                                                                                "Invalid filter in the request"
// @Failure      500         {object}  infra.ErrorResponse                                                                                                                                                "Internal server error"
// @Router       /jobs [get]
func (c *JobsController) search(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	filters, err := formatters.FormatJobFilterRequest(request)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	jobRes, err := c.ucs.Search().Execute(ctx, filters, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	response := []*api.JobResponse{}
	for _, job := range jobRes {
		response = append(response, formatters.FormatJobResponse(job))
	}

	_ = json.NewEncoder(rw).Encode(response)
}

// @Summary      Creates a new Job
// @Description  Creates a new job as part of an already created schedule
// @Tags         Jobs
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.CreateJobRequest  true  "Job creation request"
// @Success      200      {object}  api.JobResponse                                                                                                                                        "Created Job"
// @Failure      400      {object}  infra.ErrorResponse                                                                                                                                 "Invalid request"
// @Failure      422      {object}  infra.ErrorResponse                                                                                                                                 "Unprocessable parameters were sent"
// @Failure      500      {object}  infra.ErrorResponse                                                                                                                                 "Internal server error"
// @Router       /jobs [post]
func (c *JobsController) create(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	jobRequest := &api.CreateJobRequest{}
	err := infra.UnmarshalBody(request.Body, jobRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if err = jobRequest.Annotations.Validate(); err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	job := formatters.FormatJobCreateRequest(jobRequest)
	jobRes, err := c.ucs.Create().Execute(ctx, job, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatJobResponse(jobRes))
}

// @Summary      Fetch a job by uuid
// @Description  Fetch a single job by uuid
// @Tags         Jobs
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        uuid  path      string                                                                                                        true  "UUID of the job"
// @Success      200   {object}  api.JobResponse  "Job found"
// @Failure      404   {object}  infra.ErrorResponse                                                                                        "Job not found"
// @Failure      500   {object}  infra.ErrorResponse                                                                                        "Internal server error"
// @Router       /jobs/{uuid} [get]
func (c *JobsController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	uuid := mux.Vars(request)["uuid"]

	jobRes, err := c.ucs.Get().Execute(ctx, uuid, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatJobResponse(jobRes))
}

// @Summary      Start a Job by UUID
// @Description  Starts a specific job by UUID, effectively executing the transaction asynchronously
// @Tags         Jobs
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        uuid  path  string  true  "UUID of the job"
// @Success      202
// @Failure      404  {object}  infra.ErrorResponse  "Job not found"
// @Failure      500  {object}  infra.ErrorResponse  "Internal server error"
// @Router       /jobs/{uuid}/start [put]
func (c *JobsController) start(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	jobUUID := mux.Vars(request)["uuid"]
	err := c.ucs.Start().Execute(ctx, jobUUID, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
}

// @Summary      Resend Job transaction by UUID
// @Description  Resend transaction of specific job by UUID, effectively executing the re-sending of transaction asynchronously
// @Tags         Jobs
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        uuid  path  string  true  "UUID of the job"
// @Success      202
// @Failure      404  {object}  infra.ErrorResponse  "Job not found"
// @Failure      500  {object}  infra.ErrorResponse  "Internal server error"
// @Router       /jobs/{uuid}/resend [put]
func (c *JobsController) resend(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	jobUUID := mux.Vars(request)["uuid"]
	err := c.ucs.ResendTx().Execute(ctx, jobUUID, multitenancy.UserInfoValue(ctx))
	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
}

// @Summary      Update job by UUID
// @Description  Update a specific job by UUID
// @Description  WARNING: Reserved for advanced users. Orchestrate does not recommend using this endpoint.
// @Tags         Jobs
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Security     JWTAuth
// @Param        request  body      api.UpdateJobRequest  true  "Job update request"
// @Success      200      {object}  api.JobResponse                                                                                                                                        "Job found"
// @Failure      400      {object}  infra.ErrorResponse                                                                                                                                 "Invalid request"
// @Failure      404      {object}  infra.ErrorResponse                                                                                                                                 "Job not found"
// @Failure      409      {object}  infra.ErrorResponse                                                                                                                                 "Job in invalid state for the given status update"
// @Failure      500      {object}  infra.ErrorResponse                                                                                                                                 "Internal server error"
// @Router       /jobs/{uuid} [patch]
func (c *JobsController) update(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	jobRequest := &api.UpdateJobRequest{}
	err := infra.UnmarshalBody(request.Body, jobRequest)
	if err != nil {
		infra.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	job := formatters.FormatJobUpdateRequest(jobRequest)
	job.UUID = mux.Vars(request)["uuid"]
	jobRes, err := c.ucs.Update().Execute(ctx, job, jobRequest.Status, jobRequest.Message,
		multitenancy.UserInfoValue(ctx))

	if err != nil {
		infra.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatJobResponse(jobRes))
}
