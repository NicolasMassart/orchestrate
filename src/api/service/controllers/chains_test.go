// +build unit

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"

	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/api/business/use-cases/mocks"
	apitestdata "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

const chainsEndpoint = "/chains"

type chainsCtrlTestSuite struct {
	suite.Suite
	registerChainUC *mocks.MockRegisterChainUseCase
	getChainUC      *mocks.MockGetChainUseCase
	searchChainUC   *mocks.MockSearchChainsUseCase
	updateChainUC   *mocks.MockUpdateChainUseCase
	deleteChainUC   *mocks.MockDeleteChainUseCase
	ctx             context.Context
	userInfo        *multitenancy.UserInfo
	router          *mux.Router
}

var _ usecases.ChainUseCases = &chainsCtrlTestSuite{}

func (s *chainsCtrlTestSuite) Register() usecases.RegisterChainUseCase {
	return s.registerChainUC
}

func (s *chainsCtrlTestSuite) Get() usecases.GetChainUseCase {
	return s.getChainUC
}

func (s *chainsCtrlTestSuite) Search() usecases.SearchChainsUseCase {
	return s.searchChainUC
}

func (s *chainsCtrlTestSuite) Update() usecases.UpdateChainUseCase {
	return s.updateChainUC
}

func (s *chainsCtrlTestSuite) Delete() usecases.DeleteChainUseCase {
	return s.deleteChainUC
}

func TestChainsController(t *testing.T) {
	s := new(chainsCtrlTestSuite)
	suite.Run(t, s)
}

func (s *chainsCtrlTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.registerChainUC = mocks.NewMockRegisterChainUseCase(ctrl)
	s.getChainUC = mocks.NewMockGetChainUseCase(ctrl)
	s.searchChainUC = mocks.NewMockSearchChainsUseCase(ctrl)
	s.updateChainUC = mocks.NewMockUpdateChainUseCase(ctrl)
	s.deleteChainUC = mocks.NewMockDeleteChainUseCase(ctrl)
	s.userInfo = multitenancy.NewUserInfo("tenantOne", "username")
	s.ctx = multitenancy.WithUserInfo(context.Background(), s.userInfo)
	s.router = mux.NewRouter()

	controller := NewChainsController(s)
	controller.Append(s.router)
}

func (s *chainsCtrlTestSuite) TestRegister() {
	s.T().Run("should execute request successfully from latest", func(t *testing.T) {
		req := apitestdata.FakeRegisterChainRequest()
		requestBytes, _ := json.Marshal(req)
		chain := testdata.FakeChain()
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodPost, chainsEndpoint, bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		expectedChain, _ := formatters.FormatRegisterChainRequest(req)
		s.registerChainUC.EXPECT().Execute(gomock.Any(), expectedChain, s.userInfo).Return(chain, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatChainResponse(chain)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with Bad request if invalid format", func(t *testing.T) {
		req := apitestdata.FakeRegisterChainRequest()
		req.Name = ""
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPost, chainsEndpoint, bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusBadRequest, rw.Code)
	})

	s.T().Run("should fail with 500 if use case fails with an unexpected error", func(t *testing.T) {
		req := apitestdata.FakeRegisterChainRequest()
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPost, chainsEndpoint, bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.registerChainUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *chainsCtrlTestSuite) TestGetOne() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		chain := testdata.FakeChain()
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodGet, chainsEndpoint+"/chainUUID", nil).
			WithContext(s.ctx)

		s.getChainUC.EXPECT().Execute(gomock.Any(), "chainUUID", s.userInfo).Return(chain, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatChainResponse(chain)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with Internal server error if use case fails", func(t *testing.T) {
		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodGet, chainsEndpoint+"/chainUUID", nil).
			WithContext(s.ctx)

		s.getChainUC.EXPECT().Execute(gomock.Any(), "chainUUID", s.userInfo).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *chainsCtrlTestSuite) TestUpdate() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		req := apitestdata.FakeUpdateChainRequest()
		requestBytes, _ := json.Marshal(req)
		chain := testdata.FakeChain()
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodPatch, chainsEndpoint+"/chainUUID", bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		nextChain, _ := formatters.FormatUpdateChainRequest(req, "chainUUID")
		s.updateChainUC.EXPECT().
			Execute(gomock.Any(), nextChain, s.userInfo).
			Return(chain, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatChainResponse(chain)
		expectedBody, _ := json.Marshal(response)
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})

	s.T().Run("should fail with 500 if use case fails with an unexpected error", func(t *testing.T) {
		req := apitestdata.FakeUpdateChainRequest()
		requestBytes, _ := json.Marshal(req)

		rw := httptest.NewRecorder()
		httpRequest := httptest.
			NewRequest(http.MethodPatch, chainsEndpoint+"/chainUUID", bytes.NewReader(requestBytes)).
			WithContext(s.ctx)

		s.updateChainUC.EXPECT().Execute(gomock.Any(), gomock.Any(), s.userInfo).Return(nil, fmt.Errorf("error"))

		s.router.ServeHTTP(rw, httpRequest)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})
}

func (s *chainsCtrlTestSuite) TestSearch() {
	s.T().Run("should execute request successfully", func(t *testing.T) {
		chain := testdata.FakeChain()
		rw := httptest.NewRecorder()
		names := []string{"name1", "name2"}

		httpRequest := httptest.
			NewRequest(http.MethodGet, "/chains?names="+strings.Join(names, ","), nil).
			WithContext(s.ctx)

		expectedFilters := &entities.ChainFilters{Names: names}
		s.searchChainUC.EXPECT().Execute(gomock.Any(), expectedFilters, s.userInfo).Return([]*entities.Chain{chain}, nil)

		s.router.ServeHTTP(rw, httpRequest)

		response := formatters.FormatChainResponse(chain)
		expectedBody, _ := json.Marshal([]*api.ChainResponse{response})
		assert.Equal(t, string(expectedBody)+"\n", rw.Body.String())
		assert.Equal(t, http.StatusOK, rw.Code)
	})
}

func (s *chainsCtrlTestSuite) TestDelete() {
	s.T().Run("should execute verify signature request successfully", func(t *testing.T) {
		acc := testdata.FakeAccount()
		acc.Address = ethcommon.HexToAddress(inputTestAddress)
		rw := httptest.NewRecorder()

		httpRequest := httptest.
			NewRequest(http.MethodDelete, chainsEndpoint+"/chainUUID", nil).
			WithContext(s.ctx)

		s.deleteChainUC.EXPECT().Execute(gomock.Any(), "chainUUID", s.userInfo).Return(nil)

		s.router.ServeHTTP(rw, httpRequest)

		assert.Equal(t, http.StatusNoContent, rw.Code)
	})
}
