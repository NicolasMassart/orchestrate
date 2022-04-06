// +build integration

package integrationtests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	http2 "net/http"
	"strings"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	infra "github.com/consensys/orchestrate/src/infra/api"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	utils2 "github.com/consensys/orchestrate/src/infra/ethclient/utils"
	quorumkeymanager "github.com/consensys/orchestrate/src/infra/quorum-key-manager/http"
	"github.com/consensys/quorum-key-manager/src/stores/api/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

const (
	waitForEnvelopeTimeOut = 3 * time.Second
)

type txSenderEthereumTestSuite struct {
	suite.Suite
	env *IntegrationEnvironment
}

func (s *txSenderEthereumTestSuite) TestPublic() {
	signedRawTx := "0xf85380839896808252088083989680808216b4a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1e"
	txHash := "0x6621fbe1e2848446e38d99bfda159cdd83f555ae0ed7a4f3e1c3c79f7d6d74f3"

	s.T().Run("should sign and send public ethereum transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedRawTx)

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", signedRawTx)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(s.T(), err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should craft, sign and send public ethereum transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.GasPrice = nil
		jobMsg.Transaction.Gas = nil
		jobMsg.Transaction.Nonce = nil
		jobMsg.InternalData.Priority = utils.PriorityVeryHigh

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedRawTx)

		feeHistory := testdata.FakeFeeHistory(new(big.Int).SetUint64(100000))
		feeHistoryResult, _ := json.Marshal(feeHistory)
		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_feeHistory")).
			Reply(http2.StatusOK).BodyString(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":%s}", feeHistoryResult))

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_estimateGas")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x5208\"}")

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount", jobMsg.Transaction.From.String(), "pending")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", signedRawTx)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "2500100000")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should craft, sign, fail to send, and resend public ethereum transaction successfully(tx-legacy)", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.GasPrice = nil
		jobMsg.Transaction.Gas = nil
		jobMsg.Transaction.Nonce = nil
		jobMsg.Transaction.TransactionType = entities.LegacyTxType
		jobMsg.InternalData.Priority = utils.PriorityVeryLow

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedRawTx)

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_gasPrice")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x989680\"}")

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_estimateGas")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x5208\"}")

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount", jobMsg.Transaction.From.String(), "pending")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", signedRawTx)).
			Reply(http2.StatusTooManyRequests).BodyString("")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusResending, "", "6000000", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", signedRawTx)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should sign and send a public onetimekey ethereum transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.TransactionType = entities.LegacyTxType
		jobMsg.InternalData.OneTimeKey = true
		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		// IMPORTANT: As we cannot infer txHash before hand, status will be updated to WARNING
		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusWarning, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should retry on QKM client if TooManyRequests", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusTooManyRequests).JSON(&types.EthAccountResponse{})

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedRawTx)

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", signedRawTx)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should update to FAILED if key-manager fails to sign", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusBadRequest)

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should update to FAILED if rpc response with an error", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedRawTx)

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", signedRawTx)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"error\":\"invalid_raw\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should update to FAILED and don't retry if key-manager fails to sign and to update status to FAILED", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()

		url := fmt.Sprintf("/stores/%s/ethereum/accounts/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())
		gock.New(keyManagerURL).Post(url).Reply(http2.StatusUnauthorized)

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})
}

func (s *txSenderEthereumTestSuite) TestRawPublic() {
	raw := "0xf85380839896808252088083989680808216b4a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1e"
	txHash := "0x6621fbe1e2848446e38d99bfda159cdd83f555ae0ed7a4f3e1c3c79f7d6d74f3"

	s.T().Run("should send ethereum raw transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.EthereumRawTransaction
		jobMsg.Transaction.Raw = hexutil.MustDecode(raw)
		jobMsg.Transaction.Hash = utils.StringToETHHash(txHash)
		jobMsg.InternalData.OneTimeKey = true

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", raw)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should resend ethereum raw transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.EthereumRawTransaction
		jobMsg.Transaction.Raw = hexutil.MustDecode(raw)
		jobMsg.Transaction.Hash = utils.StringToETHHash(txHash)
		jobMsg.InternalData.OneTimeKey = true

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", raw)).
			Reply(429).BodyString("")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusResending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", raw)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should update to FAILED if send ethereum raw transaction fails", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.EthereumRawTransaction
		jobMsg.Transaction.Raw = hexutil.MustDecode(raw)
		jobMsg.Transaction.Hash = utils.StringToETHHash(txHash)
		jobMsg.InternalData.OneTimeKey = true

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction", raw)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"invalid_raw\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)
		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})
}

func (s *txSenderEthereumTestSuite) TestEEA() {
	txHash := "0x6621fbe1e2848446e38d99bfda159cdd83f555ae0ed7a4f3e1c3c79f7d6d74f3"
	signedRawTx := "0xf8be8080808083989680808216b4a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1ea0035695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486af842a0035695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486aa0075695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486a8a72657374726963746564"

	s.T().Run("should sign and send a EEA transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.EEAPrivateTransaction
		jobMsg.Transaction.Nonce = nil
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-eea-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedRawTx)

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "priv_getEeaTransactionCount", jobMsg.Transaction.From.String(),
				jobMsg.Transaction.PrivateFrom, jobMsg.Transaction.PrivateFor)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "priv_distributeRawTransaction", signedRawTx)).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusStored, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should sign and send EEA transaction with oneTimeKey successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.EEAPrivateTransaction
		jobMsg.InternalData.OneTimeKey = true
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "priv_distributeRawTransaction")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusStored, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should update to FAILED if we fail to send EEA transaction", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.EEAPrivateTransaction
		jobMsg.InternalData.OneTimeKey = true
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "priv_distributeRawTransaction")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"invalid_raw\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})
}

func (s *txSenderEthereumTestSuite) TestTesseraMarking() {
	signedTxRaw := "0xf851808398968082520880839896808026a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1e"
	txHash := "0x226d79b217b5ebfeddd08662f3ae1bb1b2cb339d50bbcb708b53ad5f4c71c5ea"

	s.T().Run("should sign and send Tessera marking transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.Nonce = nil
		jobMsg.Type = entities.GoQuorumMarkingTransaction
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-quorum-private-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedTxRaw)

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawPrivateTransaction", signedTxRaw, map[string]interface{}{
				"privateFor":  jobMsg.Transaction.PrivateFor,
				"privacyFlag": entities.PrivacyFlagSP,
			})).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should sign, fail to send, and resend Tessera marking transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.Nonce = nil
		jobMsg.Type = entities.GoQuorumMarkingTransaction
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-quorum-private-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedTxRaw)

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawPrivateTransaction", signedTxRaw, map[string]interface{}{
				"privateFor":  jobMsg.Transaction.PrivateFor,
				"privacyFlag": entities.PrivacyFlagSP,
			})).
			Reply(429).BodyString("")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusResending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawPrivateTransaction", signedTxRaw, map[string]interface{}{
				"privateFor":  jobMsg.Transaction.PrivateFor,
				"privacyFlag": entities.PrivacyFlagSP,
			})).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should sign and send Tessera marking transaction with oneTimeKey successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.GoQuorumMarkingTransaction
		jobMsg.Transaction.TransactionType = entities.LegacyTxType
		jobMsg.InternalData.OneTimeKey = true

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawPrivateTransaction")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		// HASH won't match
		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusWarning, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("should update to FAILED if we fail to send Tessera marking transaction", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Type = entities.GoQuorumMarkingTransaction
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(keyManagerURL).
			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
			Tags: map[string]string{
				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
			},
		})

		gock.New(keyManagerURL).
			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-quorum-private-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
			Reply(http2.StatusOK).BodyString(signedTxRaw)

		gock.New(apiURL).
			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
			AddMatcher(ethCallMatcher(wg, "eth_sendRawPrivateTransaction")).
			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"error\":\"invalid_raw\"}")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})
}

func (s *txSenderEthereumTestSuite) TestTesseraPrivate() {
	data := utils.StringToHexBytes("0xf8c380839896808252088083989680808216b4a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1ea0035695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486af842a0035695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486aa0075695b4cc4b0941e60551d7a19cf30603db5bfc23e5ac43a56f57f25f75486a8a72657374726963746564")
	enclaveKey := hexutil.MustDecode("0x226d79b217b5ebfeddd08662f3ae1bb1b2cb339d50bbcb708b53ad5f4c71c5ea")

	s.T().Run("should send a Tessera private transaction successfully", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.Data = data
		jobMsg.Type = entities.GoQuorumPrivateTransaction
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		encodedKey := base64.StdEncoding.EncodeToString(enclaveKey)
		gock.New(apiURL).
			Post(fmt.Sprintf("/tessera/%s/storeraw", jobMsg.ChainUUID)).
			Reply(http2.StatusOK).JSON(&rpc.StoreRawResponse{Key: encodedKey})

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusStored, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		jobMsg.Transaction.Nonce = nil
		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})

	s.T().Run("fail to send a Tessera private transaction", func(t *testing.T) {
		defer gock.OffAll()
		wg := &multierror.Group{}

		jobMsg := fakeMsgJob()
		jobMsg.Transaction.Data = data
		jobMsg.Type = entities.GoQuorumPrivateTransaction
		jobMsg.Transaction.TransactionType = entities.LegacyTxType

		gock.New(apiURL).
			Post(fmt.Sprintf("/tessera/%s/storeraw", jobMsg.ChainUUID)).
			Reply(http2.StatusOK).BodyString("Invalid_Response")

		gock.New(apiURL).
			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
			Reply(http2.StatusOK).JSON(&api.JobResponse{})

		err := s.sendJobMessage(jobMsg)
		require.NoError(t, err)

		err = waitTimeout(wg, waitForEnvelopeTimeOut)
		assert.NoError(t, err)
	})
}

func (s *txSenderEthereumTestSuite) TestXNonceManager() {
	signedRawTx := "0xf85380839896808252088083989680808216b4a0d35c752d3498e6f5ca1630d264802a992a141ca4b6a3f439d673c75e944e5fb0a05278aaa5fabbeac362c321b54e298dedae2d31471e432c26ea36a8d49cf08f1e"
	txHash := "0x6621fbe1e2848446e38d99bfda159cdd83f555ae0ed7a4f3e1c3c79f7d6d74f3"
	txHash2 := "0x6621fbe1e2848446e38d99bfda159cdd83f555ae0ed7a4f3e1c3c79f7d6d74f4"

	s.T().Run("should increment account nonce on consecutive transaction successfully", func(t *testing.T) {
		jobMsg := fakeMsgJob()
		jobMsg.Transaction.Nonce = nil
		defer gock.OffAll()

		for idx := 0; idx < 3; idx++ {
			wg := &multierror.Group{}
			jobMsg.UUID = uuid.Must(uuid.NewV4()).String()
			if idx == 0 {
				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount", jobMsg.Transaction.From.String(), "pending")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")
			}

			gock.New(keyManagerURL).
				Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
				Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
				Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
				Tags: map[string]string{
					quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
				},
			})

			gock.New(keyManagerURL).
				Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
				Reply(http2.StatusOK).BodyString(signedRawTx)

			if idx == 2 {
				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash2 + "\"}")
			} else {
				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")
			}

			gock.New(apiURL).
				Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
				AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, fmt.Sprintf("%d", idx), "", "")).
				Reply(http2.StatusOK).JSON(&api.JobResponse{})

			// Warning because txHash does not match
			if idx == 2 {
				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusWarning, fmt.Sprintf("%d", idx), "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})
			}

			err := s.sendJobMessage(jobMsg)
			if err != nil {
				assert.Fail(t, err.Error())
				return
			}

			err = waitTimeout(wg, waitForEnvelopeTimeOut)
			assert.NoError(t, err)
			gock.Off()
		}

		nonce, err := s.env.ns.GetLastSent(jobMsg.PartitionKey())
		require.NoError(t, err)
		assert.Equal(t, uint64(2), nonce)
	})

	s.T().Run("should re-fetch nonce on nonce too low errors", func(t *testing.T) {
		jobMsg := fakeMsgJob()
		jobMsg.Transaction.Nonce = nil
		defer gock.OffAll()

		for idx := 0; idx < 3; idx++ {
			wg := &multierror.Group{}
			jobMsg.UUID = uuid.Must(uuid.NewV4()).String()

			if idx == 0 {
				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount", jobMsg.Transaction.From.String(), "pending")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")

				gock.New(keyManagerURL).
					Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
					Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
					Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
					Tags: map[string]string{
						quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
					},
				})

				gock.New(keyManagerURL).
					Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
					Reply(http2.StatusOK).BodyString(signedRawTx)

				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, fmt.Sprintf("%d", idx), "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})
			}

			if idx == 1 {
				gock.New(keyManagerURL).
					Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).Times(2).
					Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
					Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
					Tags: map[string]string{
						quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
					},
				})

				gock.New(keyManagerURL).
					Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).Times(2).
					Reply(http2.StatusOK).BodyString(signedRawTx)

				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "1", "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})

				resp := utils2.JSONRpcMessage{Error: &utils2.JSONError{Code: 100, Message: "nonce too low"}}
				bresp, _ := json.Marshal(resp)
				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
					Reply(http2.StatusOK).BodyString(string(bresp))

				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount", jobMsg.Transaction.From.String(), "pending")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x2\"}")

				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusRecovering, "", "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})

				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash + "\"}")

				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "2", "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})
			}

			if idx > 1 {
				gock.New(keyManagerURL).
					Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
					Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
					Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
					Tags: map[string]string{
						quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
					},
				})

				gock.New(keyManagerURL).
					Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
					Reply(http2.StatusOK).BodyString(signedRawTx)

				gock.New(apiURL).
					Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
					AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
					Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"" + txHash2 + "\"}")

				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "3", "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})

				gock.New(apiURL).
					Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
					AddMatcher(txStatusUpdateMatcher(wg, entities.StatusWarning, "3", "", "")).
					Reply(http2.StatusOK).JSON(&api.JobResponse{})
			}

			err := s.sendJobMessage(jobMsg)
			if err != nil {
				assert.Fail(t, err.Error())
				return
			}

			err = waitTimeout(wg, waitForEnvelopeTimeOut)
			assert.NoError(t, err)
			gock.Off()
		}

		nonce, err := s.env.ns.GetLastSent(jobMsg.PartitionKey())
		require.NoError(t, err)
		assert.Equal(t, uint64(3), nonce)
	})

	// s.T().Run("should retry on nonce too low errors till max recover", func(t *testing.T) {
	// 	jobMsg := fakeMsgJob()
	// 	jobMsg.Transaction.Nonce = nil
	//
	// 	for idx := 0; idx <= maxRecoveryDefault; idx++ {
	// 		wg := &multierror.Group{}
	//
	// 		gock.New(apiURL).
	// 			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
	// 			AddMatcher(ethCallMatcher(wg, "eth_getTransactionCount", jobMsg.Transaction.From.String(), "pending")).
	// 			Reply(http2.StatusOK).BodyString("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":\"0x0\"}")
	//
	// 		gock.New(keyManagerURL).
	// 			Get(fmt.Sprintf("/stores/%s/ethereum/%s", qkmStoreName, jobMsg.Transaction.From.String())).
	// 			Reply(http2.StatusOK).JSON(&types.EthAccountResponse{
	// 			Address: ethcommon.HexToAddress(jobMsg.Transaction.From.String()),
	// 			Tags: map[string]string{
	// 				quorumkeymanager.TagIDAllowedTenants: multitenancy.DefaultTenant,
	// 			},
	// 		})
	//
	// 		gock.New(keyManagerURL).
	// 			Post(fmt.Sprintf("/stores/%s/ethereum/%s/sign-transaction", qkmStoreName, jobMsg.Transaction.From.String())).
	// 			Reply(http2.StatusOK).BodyString(signedRawTx)
	//
	// 		gock.New(apiURL).
	// 			Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
	// 			AddMatcher(txStatusUpdateMatcher(wg, entities.StatusPending, "0", "", "")).
	// 			Reply(http2.StatusOK).JSON(&api.JobResponse{})
	//
	// 		resp := utils2.JSONRpcMessage{Error: &utils2.JSONError{Code: 100, Message: "nonce too low"}}
	// 		bresp, _ := json.Marshal(resp)
	// 		gock.New(apiURL).
	// 			Post(fmt.Sprintf("/proxy/chains/%s", jobMsg.ChainUUID)).
	// 			AddMatcher(ethCallMatcher(wg, "eth_sendRawTransaction")).
	// 			Reply(http2.StatusOK).BodyString(string(bresp))
	//
	// 		if idx < maxRecoveryDefault {
	// 			gock.New(apiURL).
	// 				Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
	// 				AddMatcher(txStatusUpdateMatcher(wg, entities.StatusRecovering, "", "", "")).
	// 				Reply(http2.StatusOK).JSON(&api.JobResponse{})
	// 		} else {
	// 			gock.New(apiURL).
	// 				Patch(fmt.Sprintf("/jobs/%s", jobMsg.UUID)).
	// 				AddMatcher(txStatusUpdateMatcher(wg, entities.StatusFailed, "", "", "")).
	// 				Reply(http2.StatusOK).JSON(&api.JobResponse{})
	// 		}
	//
	// 		if idx == 0 {
	// 			err := s.sendJobMessage(jobMsg)
	// 			if err != nil {
	// 				assert.Fail(t, err.Error())
	// 				return
	// 			}
	// 		}
	//
	// 		err := waitTimeout(wg, waitForEnvelopeTimeOut)
	// 		assert.NoError(t, err)
	//
	// 		gock.OffAll()
	// 	}
	//
	// 	_, err := s.env.ns.GetLastSent(envelope.PartitionKey())
	// 	require.True(t, errors.IsNotFoundError(err))
	// })
}

func (s *txSenderEthereumTestSuite) TestZHealthCheck() {
	type healthRes struct {
		API   string `json:"api,omitempty"`
		Kafka string `json:"kafka,omitempty"`
		Redis string `json:"redis,omitempty"`
	}

	httpClient := http.NewClient(http.NewDefaultConfig())
	ctx := s.env.ctx
	s.T().Run("should retrieve positive health check over service dependencies", func(t *testing.T) {
		req, err := http2.NewRequest("GET", fmt.Sprintf("%s/ready?full=1", s.env.metricsURL), nil)
		assert.NoError(s.T(), err)

		gock.New(apiMetricsURL).Get("/live").Reply(http2.StatusOK)
		defer gock.OffAll()

		resp, err := httpClient.Do(req)
		if err != nil {
			assert.Fail(s.T(), err.Error())
			return
		}

		assert.Equal(s.T(), 200, resp.StatusCode)
		status := healthRes{}
		err = infra.UnmarshalBody(resp.Body, &status)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), "OK", status.API)
		assert.Equal(s.T(), "OK", status.Kafka)
		assert.Equal(s.T(), "OK", status.Redis)
	})

	s.T().Run("should retrieve a negative health check over kafka service", func(t *testing.T) {
		req, err := http2.NewRequest("GET", fmt.Sprintf("%s/ready?full=1", s.env.metricsURL), nil)
		assert.NoError(s.T(), err)

		gock.New(apiMetricsURL).Get("/live").Reply(http2.StatusOK)
		defer gock.OffAll()

		// Kill Kafka on first call so data is added in Postgres and status is CREATED but does not get updated to STARTED
		err = s.env.client.Stop(ctx, kafkaContainerID)
		assert.NoError(t, err)

		resp, err := httpClient.Do(req)
		if err != nil {
			assert.Fail(s.T(), err.Error())
			return
		}

		err = s.env.client.StartServiceAndWait(ctx, kafkaContainerID, 10*time.Second)
		assert.NoError(t, err)

		assert.Equal(s.T(), 503, resp.StatusCode)
		status := healthRes{}
		err = infra.UnmarshalBody(resp.Body, &status)
		assert.NoError(s.T(), err)
		assert.NotEqual(s.T(), "OK", status.Kafka)
		assert.Equal(s.T(), "OK", status.API)
		assert.Equal(s.T(), "OK", status.Redis)
	})
}

func fakeMsgJob() *entities.Job {
	job := testdata.FakeJob()
	job.Type = entities.EthereumTransaction
	job.Transaction.Nonce = utils.ToPtr(uint64(0)).(*uint64)
	job.Transaction.Data = nil
	job.Transaction.Raw = nil
	job.Transaction.Gas = utils.ToPtr(uint64(21000)).(*uint64)
	job.Transaction.GasPrice = utils.HexToBigInt("0x989680")
	job.Transaction.Value = utils.HexToBigInt("0x989680")
	job.Transaction.PrivateFrom = "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="
	job.Transaction.PrivateFor = []string{"B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="}
	job.InternalData.StoreID = qkmStoreName
	job.InternalData.ChainID, _ = new(big.Int).SetString("2888", 10)
	job.TenantID = "tenantID"
	return job
}

func (s *txSenderEthereumTestSuite) sendJobMessage(job *entities.Job) error {
	s.env.logger.WithField("job", job.UUID).WithField("topic", s.env.txSenderCfg.ConsumerTopic).
		Info("sending message")
	err := s.env.producer.SendJobMessage(s.env.txSenderCfg.ConsumerTopic, job, multitenancy.NewInternalAdminUser())
	if err != nil {
		return err
	}

	return nil
}

func waitTimeout(wg *multierror.Group, duration time.Duration) error {
	c := make(chan bool, 1)
	var err error
	go func() {
		defer close(c)
		err = wg.Wait().ErrorOrNil()
	}()

	select {
	case <-c:
		return err
	case <-time.After(duration):
		return fmt.Errorf("timeout after %s", duration.String())
	}
}

func txStatusUpdateMatcher(wg *multierror.Group, status entities.JobStatus, nonce string, gasPrice string, gasFeeCap string) gock.MatchFunc {
	cerr := make(chan error, 1)
	wg.Go(func() error {
		return <-cerr
	})

	return func(rw *http2.Request, _ *gock.Request) (bool, error) {
		defer func() {
			cerr <- nil
		}()

		body, _ := ioutil.ReadAll(rw.Body)
		rw.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		req := api.UpdateJobRequest{}
		if err := json.Unmarshal(body, &req); err != nil {
			cerr <- err
			return false, err
		}

		if status != "" && req.Status != status {
			err := fmt.Errorf("invalid status, got %s, expected %s", req.Status, status)
			cerr <- err
			return false, err
		}
		if nonce != "" && fmt.Sprintf("%d", *req.Transaction.Nonce) != nonce {
			err := fmt.Errorf("invalid nonce, got %d, expected %s", *req.Transaction.Nonce, nonce)
			cerr <- err
			return false, err
		}
		if gasPrice != "" && req.Transaction.GasPrice.ToInt().String() != gasPrice {
			err := fmt.Errorf("invalid gasPrice, got %s, expected %s", req.Transaction.GasPrice.ToInt().String(), gasPrice)
			cerr <- err
			return false, err
		}
		if gasFeeCap != "" && req.Transaction.GasFeeCap.ToInt().String() != gasFeeCap {
			err := fmt.Errorf("invalid gasFeeCap, got %s, expected %s", req.Transaction.GasFeeCap.ToInt().String(), gasFeeCap)
			cerr <- err
			return false, err
		}

		return true, nil
	}
}

func ethCallMatcher(wg *multierror.Group, method string, args ...interface{}) gock.MatchFunc {
	cerr := make(chan error, 1)
	wg.Go(func() error {
		select {
		case err := <-cerr:
			return err
		case <-time.After(waitForEnvelopeTimeOut):
			return fmt.Errorf("timeout after %s", waitForEnvelopeTimeOut.String())
		}
	})

	return func(rw *http2.Request, grw *gock.Request) (bool, error) {
		defer func() {
			cerr <- nil
		}()

		body, _ := ioutil.ReadAll(rw.Body)
		rw.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		req := utils2.JSONRpcMessage{}
		if err := json.Unmarshal(body, &req); err != nil {
			cerr <- err
			return false, err
		}

		if req.Method != method {
			err := fmt.Errorf("invalid method, got %s, expected %s", req.Method, method)
			cerr <- err
			return false, err
		}

		if len(args) > 0 {
			params, _ := json.Marshal(args)
			if strings.ToLower(string(req.Params)) != strings.ToLower(string(params)) {
				err := fmt.Errorf("invalid params, got %s, expected %s", req.Params, params)
				cerr <- err
				return false, err
			}
		}

		return true, nil
	}
}
