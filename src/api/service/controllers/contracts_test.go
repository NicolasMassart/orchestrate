// +build unit

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities/testdata"
	apitestdata "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/pkg/utils"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type contractsCtrlTestSuite struct {
	suite.Suite
	getContractsCatalog   *mocks.MockGetContractsCatalogUseCase
	getContract           *mocks.MockGetContractUseCase
	getContractEvents     *mocks.MockGetContractEventsUseCase
	getContractTags       *mocks.MockGetContractTagsUseCase
	registerContractEvent *mocks.MockRegisterContractDeploymentUseCase
	registerContract      *mocks.MockRegisterContractUseCase
	searchContract        *mocks.MockSearchContractUseCase
	router                *mux.Router
}

var _ usecases.ContractUseCases = &contractsCtrlTestSuite{}

func (s *contractsCtrlTestSuite) GetContractsCatalog() usecases.GetContractsCatalogUseCase {
	return s.getContractsCatalog
}
func (s *contractsCtrlTestSuite) GetContract() usecases.GetContractUseCase {
	return s.getContract
}
func (s *contractsCtrlTestSuite) GetContractEvents() usecases.GetContractEventsUseCase {
	return s.getContractEvents
}
func (s *contractsCtrlTestSuite) GetContractTags() usecases.GetContractTagsUseCase {
	return s.getContractTags
}
func (s *contractsCtrlTestSuite) SetContractCodeHash() usecases.RegisterContractDeploymentUseCase {
	return s.registerContractEvent
}
func (s *contractsCtrlTestSuite) RegisterContract() usecases.RegisterContractUseCase {
	return s.registerContract
}
func (s *contractsCtrlTestSuite) SearchContract() usecases.SearchContractUseCase {
	return s.searchContract
}

func TestContractController(t *testing.T) {
	s := new(contractsCtrlTestSuite)
	suite.Run(t, s)
}

func (s *contractsCtrlTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.getContractsCatalog = mocks.NewMockGetContractsCatalogUseCase(ctrl)
	s.getContract = mocks.NewMockGetContractUseCase(ctrl)
	s.getContractEvents = mocks.NewMockGetContractEventsUseCase(ctrl)
	s.getContractTags = mocks.NewMockGetContractTagsUseCase(ctrl)
	s.registerContractEvent = mocks.NewMockRegisterContractDeploymentUseCase(ctrl)
	s.registerContract = mocks.NewMockRegisterContractUseCase(ctrl)
	s.searchContract = mocks.NewMockSearchContractUseCase(ctrl)
	s.router = mux.NewRouter()

	controller := NewContractsController(s)
	controller.Append(s.router)
}

func (s *contractsCtrlTestSuite) TestContractsController_Register() {
	ctx := context.Background()
	s.T().Run("should execute register contract request successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeRegisterContractRequest()
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, "/contracts", bytes.NewReader(requestBytes)).
			WithContext(ctx)

		expectedContract, _ := formatters.FormatRegisterContractRequest(req)
		s.registerContract.EXPECT().Execute(gomock.Any(), expectedContract).Return(nil)

		contract := testdata.FakeContract()
		s.getContract.EXPECT().Execute(gomock.Any(), req.Name, req.Tag).Return(contract, nil)

		s.router.ServeHTTP(rw, httpRequest)
		expectedBody, _ := json.Marshal(formatters.FormatContractResponse(contract))
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
	})

	s.T().Run("should fail to register contract request if invalid format", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeRegisterContractRequest()
		req.Name = ""
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, "/contracts", bytes.NewReader(requestBytes)).
			WithContext(ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail to register contract if register contract fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeRegisterContractRequest()
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, "/contracts", bytes.NewReader(requestBytes)).
			WithContext(ctx)

		expectedContract, _ := formatters.FormatRegisterContractRequest(req)
		s.registerContract.EXPECT().Execute(gomock.Any(), expectedContract).Return(fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	s.T().Run("should fail to register contract if get contract fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeRegisterContractRequest()
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, "/contracts", bytes.NewReader(requestBytes)).
			WithContext(ctx)

		expectedContract, _ := formatters.FormatRegisterContractRequest(req)
		s.registerContract.EXPECT().Execute(gomock.Any(), expectedContract).Return(nil)

		s.getContract.EXPECT().Execute(gomock.Any(), req.Name, req.Tag).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *contractsCtrlTestSuite) TestContractsController_CodeHash() {
	ctx := context.Background()
	chainID := "2017"
	address := testdata.FakeAddress()

	s.T().Run("should execute set contract codeHash successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeSetContractCodeHashRequest()
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, fmt.Sprintf("/contracts/accounts/%s/%s", chainID, address.Hex()), bytes.NewReader(requestBytes)).
			WithContext(ctx)

		s.registerContractEvent.EXPECT().
			Execute(gomock.Any(), chainID, *address, req.CodeHash).
			Return(nil)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, "OK", rw.Body.String())
	})

	s.T().Run("should fail set contract codeHash if address is not valid", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeSetContractCodeHashRequest()
		address2 := "invalid_address_2"
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, fmt.Sprintf("/contracts/accounts/%s/%s", chainID, address2), bytes.NewReader(requestBytes)).
			WithContext(ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail set contract codeHash if set contract usecase fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := apitestdata.FakeSetContractCodeHashRequest()
		requestBytes, _ := json.Marshal(req)
		httpRequest := httptest.
			NewRequest(http.MethodPost, fmt.Sprintf("/contracts/accounts/%s/%s", chainID, address), bytes.NewReader(requestBytes)).
			WithContext(ctx)

		s.registerContractEvent.EXPECT().
			Execute(gomock.Any(), chainID, *address, req.CodeHash).
			Return(fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *contractsCtrlTestSuite) TestContractsController_GetContract() {
	ctx := context.Background()

	s.T().Run("should execute get contract successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		contract := testdata.FakeContract()
		httpRequest := httptest.
			NewRequest(http.MethodGet, fmt.Sprintf("/contracts/%s/%s", contract.Name, contract.Tag), nil).
			WithContext(ctx)

		s.getContract.EXPECT().Execute(gomock.Any(), contract.Name, contract.Tag).Return(contract, nil)

		s.router.ServeHTTP(rw, httpRequest)
		expectedBody, _ := json.Marshal(formatters.FormatContractResponse(contract))
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
	})

	s.T().Run("should fail to get contract if usecase fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		contract := testdata.FakeContract()
		httpRequest := httptest.
			NewRequest(http.MethodGet, fmt.Sprintf("/contracts/%s/%s", contract.Name, contract.Tag), nil).
			WithContext(ctx)

		s.getContract.EXPECT().Execute(gomock.Any(), contract.Name, contract.Tag).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *contractsCtrlTestSuite) TestContractsController_SearchContract() {
	ctx := context.Background()

	req := apitestdata.FakeSearchContractRequest()

	s.T().Run("should execute search contract by code_hash successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		contract := testdata.FakeContract()
		httpRequest := httptest.
			NewRequest(http.MethodGet, fmt.Sprintf("/contracts/search?code_hash=%s", req.CodeHash.String()), nil).
			WithContext(ctx)

		s.searchContract.EXPECT().Execute(gomock.Any(), req.CodeHash, nil).Return(contract, nil)

		s.router.ServeHTTP(rw, httpRequest)
		expectedBody, _ := json.Marshal(formatters.FormatContractResponse(contract))
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
	})

	s.T().Run("should execute search contract by address successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		contract := testdata.FakeContract()
		httpRequest := httptest.
			NewRequest(http.MethodGet, fmt.Sprintf("/contracts/search?address=%s", req.Address.String()), nil).
			WithContext(ctx)

		s.searchContract.EXPECT().Execute(gomock.Any(), nil, req.Address).Return(contract, nil)

		s.router.ServeHTTP(rw, httpRequest)
		expectedBody, _ := json.Marshal(formatters.FormatContractResponse(contract))
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
	})
}

func (s *contractsCtrlTestSuite) TestContractsController_GetContractEvents() {
	ctx := context.Background()
	address := ethcommon.HexToAddress(utils.RandHexString(10))
	sigHash := utils.StringToHexBytes("0x" + utils.RandHexString(10))
	indexInput := uint32(2)
	chainID := "2017"

	event := testdata.FakeEventABI()
	defaultEvent := testdata.FakeEventABI()
	rawEvent, _ := json.Marshal(event)
	rawDefaultEvent, _ := json.Marshal(defaultEvent)

	s.T().Run("should execute get contract events successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodGet, fmt.Sprintf("/contracts/accounts/%s/%s/events?sig_hash=%s&indexed_input_count=%d",
				chainID, address.Hex(), sigHash, indexInput), nil).
			WithContext(ctx)

		s.getContractEvents.EXPECT().Execute(gomock.Any(), chainID, address, sigHash, indexInput).Return(string(rawEvent), []string{string(rawDefaultEvent)}, nil)

		s.router.ServeHTTP(rw, httpRequest)
		expectedBody, _ := json.Marshal(api.GetContractEventsBySignHashResponse{Event: string(rawEvent), DefaultEvents: []string{string(rawDefaultEvent)}})
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
	})

	s.T().Run("should failt get contract events if address is invalid", func(t *testing.T) {
		invalidAddr := "invalid_address"
		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodGet, fmt.Sprintf("/contracts/accounts/%s/%s/events?sig_hash=%s&indexed_input_count=%d",
				chainID, invalidAddr, sigHash, indexInput), nil).
			WithContext(ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})
}

func (s *contractsCtrlTestSuite) TestContractsController_GetContractsCatalog() {
	ctx := context.Background()

	s.T().Run("should execute get catalog successfully", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodGet, "/contracts", nil).
			WithContext(ctx)

		catalog := []string{"contractOne", "contractTwo"}
		s.getContractsCatalog.EXPECT().Execute(gomock.Any()).Return(catalog, nil)

		s.router.ServeHTTP(rw, httpRequest)
		expectedBody, _ := json.Marshal(catalog)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
	})
}
