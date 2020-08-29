package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	jsonutils "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/encoding/json"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/http/httputil"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/multitenancy"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/utils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/service/formatters"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/use-cases/transactions"
)

const (
	IdempotencyKeyHeader = "X-Idempotency-Key"
)

type TransactionsController struct {
	ucs transactions.UseCases
}

func NewTransactionsController(ucs transactions.UseCases) *TransactionsController {
	return &TransactionsController{
		ucs: ucs,
	}
}

// Add routes to router
func (c *TransactionsController) Append(router *mux.Router) {
	router.Methods(http.MethodPost).Path("/transactions/send").
		Handler(idempotencyKeyMiddleware(http.HandlerFunc(c.send)))
	router.Methods(http.MethodPost).Path("/transactions/send-raw").
		Handler(idempotencyKeyMiddleware(http.HandlerFunc(c.sendRaw)))
	router.Methods(http.MethodPost).Path("/transactions/transfer").
		Handler(idempotencyKeyMiddleware(http.HandlerFunc(c.transfer)))
	router.Methods(http.MethodPost).Path("/transactions/deploy-contract").
		Handler(idempotencyKeyMiddleware(http.HandlerFunc(c.deployContract)))
	router.Methods(http.MethodGet).Path("/transactions/{uuid}").
		Handler(http.HandlerFunc(c.getOne))
	router.Methods(http.MethodGet).Path("/transactions").
		Handler(http.HandlerFunc(c.search))
}

func idempotencyKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(IdempotencyKeyHeader) == "" {
			r.Header.Set(IdempotencyKeyHeader, utils.RandomString(16))
		}

		next.ServeHTTP(w, r)
	})
}

// @Summary Creates and sends a new contract transaction
// @Description Creates and executes a new smart contract transaction request
// @Description The transaction can be private (Tessera, Orion).
// @Description The transaction can be a One Time Key transaction in 0 gas private networks
// @Tags Transactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security JWTAuth
// @Param request body types.SendTransactionRequest{params=types.TransactionParams{gasPricePolicy=types.GasPriceParams}} true "Contract transaction request"
// @Success 202 {object} types.TransactionResponse{params=types.ETHTransactionParams,schedule=types.ScheduleResponse} "Created contract transaction request"
// @Failure 400 {object} httputil.ErrorResponse "Invalid request"
// @Failure 409 {object} httputil.ErrorResponse "Already existing transaction"
// @Failure 422 {object} httputil.ErrorResponse "Unprocessable parameters were sent"
// @Failure 500 {object} httputil.ErrorResponse "Internal server error"
// @Router /transactions/send [post]
func (c *TransactionsController) send(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	txRequest := &types.SendTransactionRequest{}
	if err := jsonutils.UnmarshalBody(request.Body, txRequest); err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if err := txRequest.Params.Validate(); err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	txReq := formatters.FormatSendTxRequest(txRequest, request.Header.Get(IdempotencyKeyHeader))
	txResponse, err := c.ucs.SendContractTransaction().Execute(ctx, txReq, multitenancy.TenantIDFromContext(ctx))
	if err != nil {
		httputil.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(rw).Encode(formatters.FormatTxResponse(txResponse))
}

// @Summary Creates and sends a new contract deployment
// @Description Creates and executes a new contract deployment request
// @Description The transaction can be private (Tessera, Orion).
// @Description The transaction can be a One Time Key transaction in 0 gas private networks
// @Tags Transactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security JWTAuth
// @Param request body types.DeployContractRequest{params=types.DeployContractParams{gasPricePolicy=types.GasPriceParams}} true "Deployment transaction request"
// @Success 202 {object} types.TransactionResponse{params=types.ETHTransactionParams,schedule=types.ScheduleResponse} "Created deployment transaction request"
// @Failure 400 {object} httputil.ErrorResponse "Invalid request"
// @Failure 409 {object} httputil.ErrorResponse "Already existing transaction"
// @Failure 422 {object} httputil.ErrorResponse "Unprocessable parameters were sent"
// @Failure 500 {object} httputil.ErrorResponse "Internal server error"
// @Router /transactions/deploy-contract [post]
func (c *TransactionsController) deployContract(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	txRequest := &types.DeployContractRequest{}
	if err := jsonutils.UnmarshalBody(request.Body, txRequest); err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if err := txRequest.Params.Validate(); err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	txReq := formatters.FormatDeployContractRequest(txRequest, request.Header.Get(IdempotencyKeyHeader))
	txResponse, err := c.ucs.SendDeployTransaction().Execute(ctx, txReq, multitenancy.TenantIDFromContext(ctx))
	if err != nil {
		httputil.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(rw).Encode(formatters.FormatTxResponse(txResponse))
}

// @Summary Creates and sends a raw transaction
// @Description Creates and executes a new raw transaction request
// @Tags Transactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security JWTAuth
// @Param request body types.RawTransactionRequest{params=types.RawTransactionParams{}} true "Raw transaction request"
// @Success 202 {object} types.TransactionResponse{params=types.ETHTransactionParams,schedule=types.ScheduleResponse} "Created raw transaction request"
// @Failure 400 {object} httputil.ErrorResponse "Invalid request"
// @Failure 409 {object} httputil.ErrorResponse "Already existing transaction"
// @Failure 422 {object} httputil.ErrorResponse "Unprocessable parameters were sent"
// @Failure 500 {object} httputil.ErrorResponse "Internal server error"
// @Router /transactions/send-raw [post]
func (c *TransactionsController) sendRaw(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	txRequest := &types.RawTransactionRequest{}
	err := jsonutils.UnmarshalBody(request.Body, txRequest)
	if err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	txReq := formatters.FormatSendRawRequest(txRequest, request.Header.Get(IdempotencyKeyHeader))
	txResponse, err := c.ucs.SendTransaction().Execute(ctx, txReq, "", multitenancy.TenantIDFromContext(ctx))
	if err != nil {
		httputil.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(rw).Encode(formatters.FormatTxResponse(txResponse))
}

// @Summary Creates and sends a transfer transaction
// @Description Creates and executes a new transfer request
// @Tags Transactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security JWTAuth
// @Param request body types.TransferRequest{params=types.TransferParams{gasPricePolicy=types.GasPriceParams}} true "Transfer transaction request"
// @Success 202 {object} types.TransactionResponse{params=types.ETHTransactionParams,schedule=types.ScheduleResponse} "Created transfer transaction request"
// @Failure 400 {object} httputil.ErrorResponse "Invalid request"
// @Failure 409 {object} httputil.ErrorResponse "Already existing transaction"
// @Failure 422 {object} httputil.ErrorResponse "Unprocessable parameters were sent"
// @Failure 500 {object} httputil.ErrorResponse "Internal server error"
// @Router /transactions/transfer [post]
func (c *TransactionsController) transfer(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	txRequest := &types.TransferRequest{}
	err := jsonutils.UnmarshalBody(request.Body, txRequest)
	if err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	txReq := formatters.FormatSendTransferRequest(txRequest, request.Header.Get(IdempotencyKeyHeader))
	txResponse, err := c.ucs.SendTransaction().Execute(ctx, txReq, "", multitenancy.TenantIDFromContext(ctx))
	if err != nil {
		httputil.WriteHTTPErrorResponse(rw, err)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(rw).Encode(formatters.FormatTxResponse(txResponse))
}

// @Summary Fetch a transaction request by uuid
// @Description Fetch a single transaction request by uuid
// @Tags Transactions
// @Produce json
// @Security ApiKeyAuth
// @Security JWTAuth
// @Param uuid path string true "UUID of the transaction request"
// @Success 200 {object} types.TransactionResponse{params=types.ETHTransactionParams,schedule=types.ScheduleResponse} "Transaction request found"
// @Failure 404 {object} httputil.ErrorResponse "Transaction request not found"
// @Failure 500 {object} httputil.ErrorResponse "Internal server error"
// @Router /transactions/{uuid} [get]
func (c *TransactionsController) getOne(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	uuid := mux.Vars(request)["uuid"]

	txRequest, err := c.ucs.GetTransaction().Execute(ctx, uuid, multitenancy.AllowedTenantsFromContext(ctx))
	if err != nil {
		httputil.WriteHTTPErrorResponse(rw, err)
		return
	}

	_ = json.NewEncoder(rw).Encode(formatters.FormatTxResponse(txRequest))
}

// @Summary Search transaction requests by provided filters
// @Description Get a list of filtered transaction requests
// @Tags Transactions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security JWTAuth
// @Param idempotency_keys query []string false "List of idempotency keys" collectionFormat(csv)
// @Success 200 {array} types.TransactionResponse{params=types.ETHTransactionParams,schedule=types.ScheduleResponse} "List of transaction requests found"
// @Failure 400 {object} httputil.ErrorResponse "Invalid filter in the request"
// @Failure 500 {object} httputil.ErrorResponse "Internal server error"
// @Router /transactions [get]
func (c *TransactionsController) search(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ctx := request.Context()

	filters, err := formatters.FormatTransactionsFilterRequest(request)
	if err != nil {
		httputil.WriteError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	txRequests, err := c.ucs.SearchTransactions().Execute(ctx, filters, multitenancy.AllowedTenantsFromContext(ctx))
	if err != nil {
		httputil.WriteHTTPErrorResponse(rw, err)
		return
	}

	var responses []*types.TransactionResponse
	for _, txRequest := range txRequests {
		responses = append(responses, formatters.FormatTxResponse(txRequest))
	}

	_ = json.NewEncoder(rw).Encode(responses)
}
