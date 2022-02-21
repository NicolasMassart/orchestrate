// +build unit

package controllers

import (
	"bytes"
	"context"
	"fmt"
	qkm "github.com/consensys/orchestrate/src/infra/quorum-key-manager/testutils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	mocks2 "github.com/consensys/quorum-key-manager/pkg/client/mock"

	"encoding/json"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	apitestdata "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities/testdata"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const endpoint = "/faucets"

type faucetsCtrlTestSuite struct {
	suite.Suite
	registerFaucetUC *mocks.MockRegisterFaucetUseCase
	getFaucetUC      *mocks.MockGetFaucetUseCase
	searchFaucetUC   *mocks.MockSearchFaucetsUseCase
	updateFaucetUC   *mocks.MockUpdateFaucetUseCase
	deleteFaucetUC   *mocks.MockDeleteFaucetUseCase
	keyManagerClient *mocks2.MockKeyManagerClient
	ctx              context.Context
	userInfo         *multitenancy.UserInfo
	router           *mux.Router
}

var _ usecases.FaucetUseCases = &faucetsCtrlTestSuite{}

func (s *faucetsCtrlTestSuite) RegisterFaucet() usecases.RegisterFaucetUseCase {
	return s.registerFaucetUC
}

func (s *faucetsCtrlTestSuite) GetFaucet() usecases.GetFaucetUseCase {
	return s.getFaucetUC
}

func (s *faucetsCtrlTestSuite) SearchFaucets() usecases.SearchFaucetsUseCase {
	return s.searchFaucetUC
}

func (s *faucetsCtrlTestSuite) UpdateFaucet() usecases.UpdateFaucetUseCase {
	return s.updateFaucetUC
}

func (s *faucetsCtrlTestSuite) DeleteFaucet() usecases.DeleteFaucetUseCase {
	return s.deleteFaucetUC
}

func TestFaucetsController(t *testing.T) {
	s := new(faucetsCtrlTestSuite)
	suite.Run(t, s)
}

func (s *faucetsCtrlTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.registerFaucetUC = mocks.NewMockRegisterFaucetUseCase(ctrl)
	s.getFaucetUC = mocks.NewMockGetFaucetUseCase(ctrl)
	s.searchFaucetUC = mocks.NewMockSearchFaucetsUseCase(ctrl)
	s.updateFaucetUC = mocks.NewMockUpdateFaucetUseCase(ctrl)
	s.deleteFaucetUC = mocks.NewMockDeleteFaucetUseCase(ctrl)

	s.userInfo = multitenancy.NewUserInfo("tenantOne", "username")
	s.ctx = multitenancy.WithUserInfo(context.Background(), s.userInfo)
	s.router = mux.NewRouter()

	controller := NewFaucetsController(s)
	controller.Append(s.router)
}

func (s *faucetsCtrlTestSuite) TestFaucetsController_Register() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		req := apitestdata.FakeRegisterFaucetRequest()
		requestBytes, _ := json.Marshal(req)
		faucet := testdata.FakeFaucet()
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.registerFaucetUC.EXPECT().Execute(gomock.Any(), formatters.FormatRegisterFaucetRequest(req), s.userInfo).Return(faucet, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatFaucetResponse(faucet)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		req := apitestdata.FakeRegisterFaucetRequest()
		req.ChainRule = ""
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail with 500 if use case fails with an unexpected error", func(t *testing.T) {
		req := apitestdata.FakeRegisterFaucetRequest()
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.registerFaucetUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *faucetsCtrlTestSuite) TestFaucetsController_GetOne() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		faucet := testdata.FakeFaucet()
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodGet, endpoint+"/faucetUUID", nil).
			WithContext(s.ctx)

		s.getFaucetUC.EXPECT().Execute(gomock.Any(), "faucetUUID", s.userInfo).Return(faucet, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatFaucetResponse(faucet)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with Internal server error if use case fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodGet, endpoint+"/faucetUUID", nil).
			WithContext(s.ctx)

		s.getFaucetUC.EXPECT().Execute(gomock.Any(), "faucetUUID", s.userInfo).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *faucetsCtrlTestSuite) TestFaucetsController_Update() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		req := apitestdata.FakeUpdateFaucetRequest()
		requestBytes, _ := json.Marshal(req)
		faucet := testdata.FakeFaucet()
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodPatch, endpoint+"/faucetUUID", bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.updateFaucetUC.EXPECT().
			Execute(gomock.Any(), formatters.FormatUpdateFaucetRequest(req, "faucetUUID"), s.userInfo).
			Return(faucet, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatFaucetResponse(faucet)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		req := apitestdata.FakeUpdateFaucetRequest()
		req.Cooldown = "notADuration"
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPatch, endpoint+"/faucetUUID", bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail with 500 if use case fails with an unexpected error", func(t *testing.T) {
		req := apitestdata.FakeUpdateFaucetRequest()
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPatch, endpoint+"/faucetUUID", bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.updateFaucetUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *faucetsCtrlTestSuite) TestFaucetsController_Search() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		faucet := testdata.FakeFaucet()
		rw := httptest.NewRecorder()
		names := []string{"name1", "name2"}
		chainRule := "chainRule"

		httpRequest := httptest.
			NewRequest(http.MethodGet, "/faucets?names="+strings.Join(names, ",")+"&chain_rule="+chainRule, nil).
			WithContext(s.ctx)

		expectedFilters := &entities.FaucetFilters{
			Names:     names,
			ChainRule: chainRule,
		}
		s.searchFaucetUC.EXPECT().Execute(gomock.Any(), expectedFilters, s.userInfo).Return([]*entities.Faucet{faucet}, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatFaucetResponse(faucet)
		expectedBody, _ := json.Marshal([]*api.FaucetResponse{response})
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})
}

func (s *faucetsCtrlTestSuite) TestFaucetsController_Delete() {
	s.T().Run("should execute verify signature request successfully", func(t *testing.T) {
		acc := testdata.FakeAccount()
		acc.Address = ethcommon.HexToAddress(inputTestAddress)
		rw := httptest.NewRecorder()
		request := qkm.FakeVerifyPayloadRequest()
		requestBytes, _ := json.Marshal(request)

		httpRequest := httptest.
			NewRequest(http.MethodDelete, endpoint+"/faucetUUID", bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.deleteFaucetUC.EXPECT().Execute(gomock.Any(), "faucetUUID", s.userInfo).Return(nil)

		s.router.ServeHTTP(rw, httpRequest)

		assert.Equal(t, http.StatusNoContent, rw.Code)
	})
}
