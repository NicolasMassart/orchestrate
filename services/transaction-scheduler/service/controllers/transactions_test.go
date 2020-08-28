// +build unit

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/encoding/json"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/multitenancy"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/testutils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/service/formatters"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/entities"
	testutils2 "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/testutils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/use-cases/transactions"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/use-cases/transactions/mocks"
)

type transactionsControllerTestSuite struct {
	suite.Suite
	controller            *TransactionsController
	router                *mux.Router
	sendContractTxUseCase *mocks.MockSendContractTxUseCase
	sendDeployTxUseCase   *mocks.MockSendDeployTxUseCase
	sendTxUseCase         *mocks.MockSendTxUseCase
	getTxUseCase          *mocks.MockGetTxUseCase
	searchTxsUsecase      *mocks.MockSearchTransactionsUseCase
	ctx                   context.Context
	tenantID              string
}

func (s *transactionsControllerTestSuite) SendContractTransaction() transactions.SendContractTxUseCase {
	return s.sendContractTxUseCase
}

func (s *transactionsControllerTestSuite) SendDeployTransaction() transactions.SendDeployTxUseCase {
	return s.sendDeployTxUseCase
}

func (s *transactionsControllerTestSuite) SendTransaction() transactions.SendTxUseCase {
	return s.sendTxUseCase
}

func (s *transactionsControllerTestSuite) GetTransaction() transactions.GetTxUseCase {
	return s.getTxUseCase
}

func (s *transactionsControllerTestSuite) SearchTransactions() transactions.SearchTransactionsUseCase {
	return s.searchTxsUsecase
}

var _ transactions.UseCases = &transactionsControllerTestSuite{}

func TestTransactionsController(t *testing.T) {
	s := new(transactionsControllerTestSuite)
	suite.Run(t, s)
}

func (s *transactionsControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.sendContractTxUseCase = mocks.NewMockSendContractTxUseCase(ctrl)
	s.sendDeployTxUseCase = mocks.NewMockSendDeployTxUseCase(ctrl)
	s.sendTxUseCase = mocks.NewMockSendTxUseCase(ctrl)
	s.getTxUseCase = mocks.NewMockGetTxUseCase(ctrl)
	s.searchTxsUsecase = mocks.NewMockSearchTransactionsUseCase(ctrl)
	s.tenantID = "tenantId"
	s.ctx = context.WithValue(context.Background(), multitenancy.TenantIDKey, s.tenantID)
	s.ctx = context.WithValue(s.ctx, multitenancy.AllowedTenantsKey, []string{s.tenantID})

	s.router = mux.NewRouter()
	s.controller = NewTransactionsController(s)
	s.controller.Append(s.router)
}

func (s *transactionsControllerTestSuite) TestTransactionsController_send() {
	urlPath := "/transactions/send"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := testutils.FakeSendTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		testutils2.FakeTxRequestEntity()
		txRequestEntityResp := testutils2.FakeTxRequestEntity()

		txRequestEntity := formatters.FormatSendTxRequest(txRequest, idempotencyKey)
		s.sendContractTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, s.tenantID).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	s.T().Run("should execute request successfully without IdempotencyKeyHeader", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := testutils.FakeSendTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		testutils2.FakeTxRequestEntity()
		txRequestEntityResp := testutils2.FakeTxRequestEntity()

		s.sendContractTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), s.tenantID).
			DoAndReturn(func(ctx context.Context, txReq *entities.TxRequest, tenantID string) (*entities.TxRequest, error) {
				if txReq.IdempotencyKey == "" {
					return nil, errors.InvalidParameterError("missing required idempotencyKey")
				}
				txRequestEntityResp.IdempotencyKey = txReq.IdempotencyKey
				return txRequestEntityResp, nil
			})

		s.router.ServeHTTP(rw, httpRequest)

		_ = formatters.FormatTxResponse(txRequestEntityResp)
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := testutils.FakeSendTransactionRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.sendContractTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), s.tenantID).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := testutils.FakeSendTransactionRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestTransactionsController_deploy() {
	urlPath := "/transactions/deploy-contract"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := testutils.FakeDeployContractRequest()
		requestBytes, _ := json.Marshal(txRequest)

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		txRequestEntityResp := testutils2.FakeTxRequestEntity()

		txRequestEntity := formatters.FormatDeployContractRequest(txRequest, idempotencyKey)
		s.sendDeployTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, s.tenantID).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := testutils.FakeDeployContractRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.sendDeployTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), s.tenantID).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := testutils.FakeDeployContractRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail with 400 if request fails with InvalidParameterError for private txs", func(t *testing.T) {
		rw := httptest.NewRecorder()
		txRequest := testutils.FakeDeployContractRequest()
		txRequest.Params.PrivateFrom = "PrivateFrom"
		requestBytes, _ := json.Marshal(txRequest)

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestTransactionsController_sendRaw() {
	urlPath := "/transactions/send-raw"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := testutils.FakeSendRawTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		txRequestEntityResp := testutils2.FakeTxRequestEntity()

		txRequestEntity := formatters.FormatSendRawRequest(txRequest, idempotencyKey)
		s.sendTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, "", s.tenantID).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := testutils.FakeSendRawTransactionRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.sendTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), "", s.tenantID).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := testutils.FakeSendRawTransactionRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestTransactionsController_transfer() {
	urlPath := "/transactions/transfer"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := testutils.FakeSendTransferTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		txRequestEntityResp := testutils2.FakeTransferTxRequestEntity()

		txRequestEntity := formatters.FormatSendTransferRequest(txRequest, idempotencyKey)
		s.sendTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, "", s.tenantID).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := testutils.FakeSendTransferTransactionRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.sendTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), "", s.tenantID).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := testutils.FakeSendTransferTransactionRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestTransactionsController_getOne() {
	uuid := "uuid"
	urlPath := "/transactions/" + uuid

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath, nil).WithContext(s.ctx)
		txRequest := testutils2.FakeTransferTxRequestEntity()

		s.getTxUseCase.EXPECT().Execute(gomock.Any(), uuid, []string{s.tenantID}).
			Return(txRequest, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequest)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with 404 if NotFoundError is returned", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath, nil).WithContext(s.ctx)

		s.getTxUseCase.EXPECT().Execute(gomock.Any(), uuid, []string{s.tenantID}).
			Return(nil, errors.NotFoundError(""))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusNotFound, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestTransactionsController_search() {
	urlPath := "/transactions"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath+"?idempotency_keys=mykey,mykey1", nil).WithContext(s.ctx)
		txRequest := testutils2.FakeTransferTxRequestEntity()
		expectedFilers := &entities.TransactionFilters{
			IdempotencyKeys: []string{"mykey", "mykey1"},
		}

		s.searchTxsUsecase.EXPECT().Execute(gomock.Any(), expectedFilers, []string{s.tenantID}).
			Return([]*entities.TxRequest{txRequest}, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := []*types.TransactionResponse{formatters.FormatTxResponse(txRequest)}
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with 400 if filer is malformed", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath+"?idempotency_keys=mykey,mykey", nil).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail with 500 if use case fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath+"?idempotency_keys=mykey,mykey1", nil).WithContext(s.ctx)
		expectedFilers := &entities.TransactionFilters{
			IdempotencyKeys: []string{"mykey", "mykey1"},
		}

		s.searchTxsUsecase.EXPECT().Execute(gomock.Any(), expectedFilers, []string{s.tenantID}).
			Return(nil, fmt.Errorf(""))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}
