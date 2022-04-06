// +build unit

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	apitypes "github.com/consensys/orchestrate/src/api/service/types"
	apitestdata "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	speedUpTxUseCase      *mocks.MockSpeedUpTxUseCase
	callOffTxUseCase      *mocks.MockCallOffTxUseCase
	ctx                   context.Context
	userInfo              *multitenancy.UserInfo
	defaultRetryInterval  time.Duration
}

func (s *transactionsControllerTestSuite) SendContract() usecases.SendContractTxUseCase {
	return s.sendContractTxUseCase
}

func (s *transactionsControllerTestSuite) SendDeploy() usecases.SendDeployTxUseCase {
	return s.sendDeployTxUseCase
}

func (s *transactionsControllerTestSuite) Send() usecases.SendTxUseCase {
	return s.sendTxUseCase
}

func (s *transactionsControllerTestSuite) Get() usecases.GetTxUseCase {
	return s.getTxUseCase
}

func (s *transactionsControllerTestSuite) Search() usecases.SearchTransactionsUseCase {
	return s.searchTxsUsecase
}

func (s *transactionsControllerTestSuite) SpeedUp() usecases.SpeedUpTxUseCase {
	return s.speedUpTxUseCase
}

func (s *transactionsControllerTestSuite) CallOff() usecases.CallOffTxUseCase {
	return s.callOffTxUseCase
}

var _ usecases.TransactionUseCases = &transactionsControllerTestSuite{}

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
	s.speedUpTxUseCase = mocks.NewMockSpeedUpTxUseCase(ctrl)
	s.callOffTxUseCase = mocks.NewMockCallOffTxUseCase(ctrl)
	s.searchTxsUsecase = mocks.NewMockSearchTransactionsUseCase(ctrl)
	s.defaultRetryInterval = time.Second * 2
	s.userInfo = multitenancy.NewUserInfo("tenantOne", "username")
	s.ctx = multitenancy.WithUserInfo(context.Background(), s.userInfo)

	s.router = mux.NewRouter()
	s.controller = NewTransactionsController(s)
	s.controller.Append(s.router)
}

func (s *transactionsControllerTestSuite) TestSend() {
	urlPath := "/transactions/send"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := apitestdata.FakeSendTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		testdata.FakeTxRequest()
		txRequestEntityResp := testdata.FakeTxRequest()

		s.sendContractTxUseCase.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	s.T().Run("should execute request successfully without IdempotencyKeyHeader", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := apitestdata.FakeSendTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		txRequestEntityResp := testdata.FakeTxRequest()

		s.sendContractTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), s.userInfo).
			DoAndReturn(func(ctx context.Context, txReq *entities.TxRequest, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error) {
				txRequestEntityResp.IdempotencyKey = txReq.IdempotencyKey
				return txRequestEntityResp, nil
			})

		s.router.ServeHTTP(rw, httpRequest)

		_ = formatters.FormatTxResponse(txRequestEntityResp)
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := apitestdata.FakeSendTransactionRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.sendContractTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), s.userInfo).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := apitestdata.FakeSendTransactionRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestDeploy() {
	urlPath := "/transactions/deploy-contract"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := apitestdata.FakeDeployContractRequest()
		requestBytes, _ := json.Marshal(txRequest)

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		txRequestEntityResp := testdata.FakeTxRequest()

		txRequestEntity := formatters.FormatDeployContractRequest(txRequest, idempotencyKey)
		s.sendDeployTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, s.userInfo).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := apitestdata.FakeDeployContractRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.sendDeployTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), s.userInfo).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := apitestdata.FakeDeployContractRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestSendRaw() {
	urlPath := "/transactions/send-raw"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := apitestdata.FakeSendRawTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		txRequestEntityResp := testdata.FakeTxRequest()

		txRequestEntity := formatters.FormatSendRawRequest(txRequest, idempotencyKey)
		s.sendTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, nil, s.userInfo).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := apitestdata.FakeSendRawTransactionRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.sendTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), nil, s.userInfo).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := apitestdata.FakeSendRawTransactionRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestTransfer() {
	urlPath := "/transactions/transfer"
	idempotencyKey := "idempotencyKey"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()

		txRequest := apitestdata.FakeSendTransferTransactionRequest()
		requestBytes, err := json.Marshal(txRequest)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}

		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)
		httpRequest.Header.Set(IdempotencyKeyHeader, idempotencyKey)

		txRequestEntityResp := testdata.FakeTransferTxRequest()

		txRequestEntity := formatters.FormatTransferRequest(txRequest, idempotencyKey)
		s.sendTxUseCase.EXPECT().Execute(gomock.Any(), txRequestEntity, nil, s.userInfo).Return(txRequestEntityResp, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequestEntityResp)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusAccepted, rw.Code)
	})

	// Sufficient test to check that the mapping to HTTP errors is working. All other status code tests are done in integration tests
	s.T().Run("should fail with 422 if use case fails with InvalidParameterError", func(t *testing.T) {
		txRequest := apitestdata.FakeSendTransferTransactionRequest()
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath,
			bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.sendTxUseCase.EXPECT().
			Execute(gomock.Any(), gomock.Any(), nil, s.userInfo).
			Return(nil, errors.InvalidParameterError("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		txRequest := apitestdata.FakeSendTransferTransactionRequest()
		txRequest.ChainName = ""
		requestBytes, _ := json.Marshal(txRequest)

		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, urlPath, bytes.NewReader(requestBytes)).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestGetOne() {
	uuid := "uuid"
	urlPath := "/transactions/" + uuid

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath, nil).WithContext(s.ctx)
		txRequest := testdata.FakeTransferTxRequest()

		s.getTxUseCase.EXPECT().Execute(gomock.Any(), uuid, s.userInfo).
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

		s.getTxUseCase.EXPECT().Execute(gomock.Any(), uuid, s.userInfo).
			Return(nil, errors.NotFoundError(""))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusNotFound, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestSpeedUp() {
	uuid := "uuid"
	increment := 0.3
	urlPath := fmt.Sprintf("/transactions/%s/speed-up?boost=%f", uuid, increment)
	txRequest := testdata.FakeTxRequest()

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPut, urlPath, nil).WithContext(s.ctx)

		s.speedUpTxUseCase.EXPECT().Execute(gomock.Any(), uuid, increment, s.userInfo).
			Return(txRequest, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequest)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with 400 if boost value is lower than min", func(t *testing.T) {
		urlPath := fmt.Sprintf("/transactions/%s/speed-up?boost=%f", uuid, 0.01)
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPut, urlPath, nil).WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail with 404 if NotFoundError is returned", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPut, urlPath, nil).WithContext(s.ctx)

		s.speedUpTxUseCase.EXPECT().Execute(gomock.Any(), uuid, increment, s.userInfo).
			Return(nil, errors.NotFoundError(""))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusNotFound, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestCallOff() {
	uuid := "uuid"
	urlPath := fmt.Sprintf("/transactions/%s/call-off", uuid)
	txRequest := testdata.FakeTxRequest()

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPut, urlPath, nil).WithContext(s.ctx)

		s.callOffTxUseCase.EXPECT().Execute(gomock.Any(), uuid, s.userInfo).
			Return(txRequest, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatTxResponse(txRequest)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with 404 if NotFoundError is returned", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPut, urlPath, nil).WithContext(s.ctx)

		s.callOffTxUseCase.EXPECT().Execute(gomock.Any(), uuid, s.userInfo).
			Return(nil, errors.NotFoundError(""))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusNotFound, rw.Code)
	})
}

func (s *transactionsControllerTestSuite) TestSearch() {
	urlPath := "/transactions"

	s.T().Run("should execute request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, urlPath+"?idempotency_keys=mykey,mykey1", nil).WithContext(s.ctx)
		txRequest := testdata.FakeTransferTxRequest()
		expectedFilers := &entities.TransactionRequestFilters{
			IdempotencyKeys: []string{"mykey", "mykey1"},
		}

		s.searchTxsUsecase.EXPECT().Execute(gomock.Any(), expectedFilers, s.userInfo).
			Return([]*entities.TxRequest{txRequest}, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := []*apitypes.TransactionResponse{formatters.FormatTxResponse(txRequest)}
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
		expectedFilers := &entities.TransactionRequestFilters{
			IdempotencyKeys: []string{"mykey", "mykey1"},
		}

		s.searchTxsUsecase.EXPECT().Execute(gomock.Any(), expectedFilers, s.userInfo).
			Return(nil, fmt.Errorf(""))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}
