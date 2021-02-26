// +build unit

package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ConsenSys/orchestrate/pkg/encoding/json"
	ethclient "github.com/ConsenSys/orchestrate/pkg/ethclient/utils"
)

func TestHTTPCacheRequest_Valid(t *testing.T) {
	msg := ethclient.JSONRpcMessage{
		Method: "eth_getTransactionReceipt",
	}
	msg.Params, _ = json.Marshal([]string{"0x7d231ca6a5fc03f5365b3d62dcfe372ed5c13ac7014d016b52ed72094919556c"})

	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewReader(body))

	c, k, ttl, err := HTTPCacheRequest(req.Context(), req)
	assert.NoError(t, err)
	assert.True(t, c)
	assert.Equal(t, time.Duration(0), ttl)
	assert.Equal(t, "eth_getTransactionReceipt([\"0x7d231ca6a5fc03f5365b3d62dcfe372ed5c13ac7014d016b52ed72094919556c\"])", k)
}

func TestHTTPCacheRequest_ValidWithCustomTTL(t *testing.T) {
	msg := ethclient.JSONRpcMessage{
		Method: "eth_getBlockByNumber",
	}
	msg.Params, _ = json.Marshal([]string{"latest"})

	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewReader(body))

	c, k, ttl, err := HTTPCacheRequest(req.Context(), req)
	assert.NoError(t, err)
	assert.True(t, c)
	assert.Equal(t, time.Second, ttl)
	assert.Equal(t, "eth_getBlockByNumber([\"latest\"])", k)
}

func TestHTTPCacheRequest_IgnoreReqType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

	c, _, _, err := HTTPCacheRequest(req.Context(), req)
	assert.NoError(t, err)
	assert.False(t, c)
}

func TestHTTPCacheRequest_IgnoreRPCMethod(t *testing.T) {
	msg := ethclient.JSONRpcMessage{
		Method: "eth_getTransactionCount",
	}

	body, _ := json.Marshal(msg)
	req := httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewReader(body))

	c, _, _, err := HTTPCacheRequest(req.Context(), req)
	assert.NoError(t, err)
	assert.False(t, c)
}
