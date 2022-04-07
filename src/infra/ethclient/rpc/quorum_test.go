// +build unit

package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/cenkalti/backoff/v4"
	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/src/infra/ethclient/testutils"
	pkgUtils "github.com/consensys/orchestrate/pkg/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func newQuorumClient() *Client {
	newBackOff := func() backoff.BackOff { return pkgUtils.NewBackOff(testutils.TestConfig) }
	return NewClientWithBackOff(newBackOff, &http.Client{
		Transport: testutils.MockRoundTripper{},
	})
}

func TestStoreRawTransaction(t *testing.T) {
	testSet := []struct {
		name                   string
		httpBodyResponse       interface{}
		httpStatusCodeResponse int
		httpErrorResponse      error
		expectedEnclaveKey     string
		expectedError          error
	}{
		{
			"success storeraw",
			StoreRawResponse{Key: base64.StdEncoding.EncodeToString(hexutil.MustDecode("0xb5eb2d"))},
			200,
			nil,
			"0xb5eb2d",
			nil,
		},
		{
			"fail storeraw",
			StoreRawResponse{Key: "test"},
			400,
			fmt.Errorf("test"),
			"0x",
			errors.HTTPConnectionError("failed to send a request to Tessera enclave: 08200@: failed to send a request to 'test/storeraw' - Post \"test/storeraw\": test"),
		},
	}
	ec := newQuorumClient()

	for _, test := range testSet {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			b, _ := json.Marshal(test.httpBodyResponse)

			ctx := testutils.NewContext(test.httpErrorResponse, test.httpStatusCodeResponse, ioutil.NopCloser(bytes.NewReader(b)))

			enclaveKey, err := ec.StoreRaw(ctx, "test", []byte{}, "testPrivateFrom")
			assert.Equal(t, err, test.expectedError)
			assert.Equal(t, test.expectedEnclaveKey, hexutil.Encode(enclaveKey))
		})
	}
}

func TestGetStatus(t *testing.T) {
	testSet := []struct {
		name                   string
		httpBodyResponse       interface{}
		httpStatusCodeResponse int
		httpErrorResponse      error
		expectedStatus         string
		expectedError          error
	}{
		{
			"success get status",
			"testStatus",
			200,
			nil,
			"\"testStatus\"",
			nil,
		},
	}
	ec := newQuorumClient()

	for _, test := range testSet {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			b, _ := json.Marshal(test.httpBodyResponse)

			ctx := testutils.NewContext(test.httpErrorResponse, test.httpStatusCodeResponse, ioutil.NopCloser(bytes.NewReader(b)))

			status, err := ec.GetStatus(ctx, "testEndpoint")
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedStatus, status)
		})
	}
}

func TestSendQuorumRawPrivateTransaction(t *testing.T) {
	ec := newQuorumClient()

	// Test 1 with Error
	ctx := testutils.NewContext(fmt.Errorf("test-error"), 0, nil)
	_, err := ec.SendQuorumRawPrivateTransaction(ctx, "test-endpoint", nil, nil, nil, 0)
	assert.Error(t, err, "#1 SendQuorumRawPrivateTransaction should  error")
}
