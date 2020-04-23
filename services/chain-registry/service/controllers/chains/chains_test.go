// +build unit

package chains

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net/http/httptest"
	"testing"

	usecases "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/chain-registry/use-cases"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store/models"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	mockethclient "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/ethereum/ethclient/mock"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/multitenancy"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store"
	mockstore "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store/mock"
)

const (
	expectedInternalServerErrorBody  = "{\"message\":\"test error\"}\n"
	expectedNotFoundErrorBody        = "{\"message\":\"DB200@: not found error\"}\n"
	expectedNotUniqueURLsError       = "{\"message\":\"42400@chain-registry.store.api: invalid body, with: field validation for 'URLs[0]' failed on the 'url' tag\"}\n"
	expectedUnknownBodyError         = "{\"message\":\"FF000@chain-registry.store.api: json: unknown field \\\"unknownField\\\"\"}\n"
	expectedSuccessStatusBody        = "{\"uuid\":\"\",\"name\":\"\",\"tenantID\":\"\",\"urls\":null,\"createdAt\":null}\n"
	expectedSuccessStatusSliceBody   = "[]\n"
	expectedSuccessStatusContentType = "application/json"
	expectedErrorStatusContentType   = "text/plain; charset=utf-8"
	expectedErrorInvalidManagerType  = "{\"message\":\"42400@chain-registry.store.api: invalid body, with: field validation for 'Type' failed on the 'isPrivateTxManagerType' tag\"}\n"
	expectedErrorInvalidManagerURL   = "{\"message\":\"42400@chain-registry.store.api: invalid body, with: field validation for 'URL' failed on the 'url' tag field validation for 'Type' failed on the 'isPrivateTxManagerType' tag\"}\n"
	notFoundErrorFilter              = "notFoundError"
)

type HTTPRouteTests struct {
	name                string
	chainAgent          func(t *testing.T) store.ChainAgent
	httpMethod          string
	path                string
	body                func() []byte
	expectedStatusCode  int
	expectedContentType string
	expectedBody        func() string
}

func TestHTTPRouteTests(t *testing.T) {
	t.Parallel()

	testsSuite := [][]HTTPRouteTests{
		deleteChainByUUIDTests,
		getChainsTests,
		getChainsByUUIDTests,
		patchChainByUUIDTests,
		postChainTests,
	}

	for _, tests := range testsSuite {
		for _, tt := range tests {
			tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel() // marks each test case as capable of running in parallel with each other

				router := mux.NewRouter()
				chainAgent := tt.chainAgent(t)
				ethClient := UseEthClient(t)
				NewController(
					usecases.NewGetChains(chainAgent),
					usecases.NewGetChain(chainAgent),
					usecases.NewRegisterChain(chainAgent, ethClient),
					usecases.NewDeleteChain(chainAgent),
					usecases.NewUpdateChain(chainAgent),
				).Append(router)

				// Normal tests
				writer := httptest.NewRecorder()
				request := httptest.NewRequest(tt.httpMethod, tt.path, bytes.NewReader(tt.body()))

				router.ServeHTTP(writer, request)
				testResponse(t, writer, tt.expectedStatusCode, tt.expectedContentType, tt.expectedBody())
			})
		}
	}

	// Multi-tenant tests
	testsSuiteMultitenant := [][]HTTPRouteTests{
		deleteChainByUUIDTests,
		getChainsTests,
		getChainsByUUIDTests,
		patchChainByUUIDTests,
		postChainTestsMultitenant,
	}

	for _, tests := range testsSuiteMultitenant {
		for _, tt := range tests {
			tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables

			t.Run(tt.name, func(t *testing.T) {
				t.Parallel() // marks each test case as capable of running in parallel with each other

				router := mux.NewRouter()
				chainAgent := tt.chainAgent(t)
				ethClient := UseEthClient(t)
				NewController(
					usecases.NewGetChains(chainAgent),
					usecases.NewGetChain(chainAgent),
					usecases.NewRegisterChain(chainAgent, ethClient),
					usecases.NewDeleteChain(chainAgent),
					usecases.NewUpdateChain(chainAgent),
				).Append(router)

				writer := httptest.NewRecorder()
				request := httptest.NewRequest(tt.httpMethod, tt.path, bytes.NewReader(tt.body()))
				request = request.WithContext(multitenancy.WithTenantID(request.Context(), "tenantID"))
				router.ServeHTTP(writer, request)
				testResponse(t, writer, tt.expectedStatusCode, tt.expectedContentType, tt.expectedBody())
			})
		}
	}

}

func testResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatusCode int, expectedContentType, expectedBody string) {
	assert.Equal(t, expectedStatusCode, w.Code, "Did not get expected HTTP status code")
	assert.Equal(t, expectedContentType, w.Header().Get("Content-Type"), "Did not get expected content type")
	assert.Equal(t, expectedBody, w.Body.String(), "Did not get expected body")
}

func UseMockChainRegistry(t *testing.T) store.ChainAgent {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mockstore.NewMockChainAgent(mockCtrl)

	mockStore.EXPECT().RegisterChain(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, chain *models.Chain) error {
			chain.UUID = "uuid"
			for idx, _ := range chain.PrivateTxManagers {
				chain.PrivateTxManagers[idx].ChainUUID = "uuid"
				chain.PrivateTxManagers[idx].UUID = "uuid"
			}
			return nil
		},
	).AnyTimes()

	mockStore.EXPECT().GetChains(gomock.Any(), gomock.Any()).Return([]*models.Chain{}, nil).AnyTimes()
	mockStore.EXPECT().GetChainsByTenant(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*models.Chain{}, nil).AnyTimes()

	mockStore.EXPECT().GetChainByUUID(gomock.Any(), gomock.Any()).Return(&models.Chain{}, nil).AnyTimes()
	mockStore.EXPECT().GetChainByUUIDAndTenant(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.Chain{}, nil).AnyTimes()

	mockStore.EXPECT().UpdateChainByName(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStore.EXPECT().UpdateChainByUUID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockStore.EXPECT().DeleteChainByUUID(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStore.EXPECT().DeleteChainByUUIDAndTenant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return mockStore
}

var errTest = fmt.Errorf("test error")
var errNotFound = errors.NotFoundError("not found error")

func UseErrorChainRegistry(t *testing.T) store.ChainAgent {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockStore := mockstore.NewMockChainAgent(mockCtrl)

	mockStore.EXPECT().RegisterChain(gomock.Any(), gomock.Any()).Return(errTest).AnyTimes()

	mockStore.EXPECT().GetChains(gomock.Any(), gomock.Any()).Return(nil, errTest).AnyTimes()
	mockStore.EXPECT().GetChainsByTenant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errTest).AnyTimes()

	mockStore.EXPECT().GetChainByUUID(gomock.Any(), gomock.Eq("0")).Return(nil, errNotFound).AnyTimes()
	mockStore.EXPECT().GetChainByUUIDAndTenant(gomock.Any(), gomock.Eq("0"), gomock.Any()).Return(nil, errNotFound).AnyTimes()

	mockStore.EXPECT().GetChainByUUID(gomock.Any(), gomock.Not(gomock.Eq("0"))).Return(nil, errTest).AnyTimes()
	mockStore.EXPECT().GetChainByUUIDAndTenant(gomock.Any(), gomock.Not(gomock.Eq("0")), gomock.Any()).Return(nil, errTest).AnyTimes()

	mockStore.EXPECT().UpdateChainByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, name string, chain *models.Chain) error {
			if name == notFoundErrorFilter {
				return errNotFound
			}
			return errTest
		}).AnyTimes()

	mockStore.EXPECT().UpdateChainByUUID(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, uuid string, chain *models.Chain) error {
			if uuid == "0" {
				return errNotFound
			}
			return errTest
		}).AnyTimes()

	mockStore.EXPECT().DeleteChainByUUID(gomock.Any(), gomock.Eq("0")).Return(errNotFound).AnyTimes()
	mockStore.EXPECT().DeleteChainByUUIDAndTenant(gomock.Any(), gomock.Eq("0"), gomock.Any()).Return(errNotFound).AnyTimes()

	mockStore.EXPECT().DeleteChainByUUID(gomock.Any(), gomock.Not(gomock.Eq("0"))).Return(errTest).AnyTimes()
	mockStore.EXPECT().DeleteChainByUUIDAndTenant(gomock.Any(), gomock.Not(gomock.Eq("0")), gomock.Any()).Return(errTest).AnyTimes()

	return mockStore
}

func UseEthClient(t *testing.T) *mockethclient.MockChainLedgerReader {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockEthClient := mockethclient.NewMockChainLedgerReader(mockCtrl)

	mockEthClient.EXPECT().HeaderByNumber(gomock.Any(), gomock.Any(), gomock.Any()).Return(&ethtypes.Header{Number: big.NewInt(666)}, nil).AnyTimes()

	return mockEthClient
}