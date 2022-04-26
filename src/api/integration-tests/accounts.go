// +build integration

package integrationtests

import (
	"net/http"
	"testing"
	"time"

	"github.com/consensys/orchestrate/pkg/ethereum/account"
	qkm "github.com/consensys/orchestrate/src/infra/quorum-key-manager/testutils"
	utilstypes "github.com/consensys/quorum-key-manager/src/utils/api/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	api "github.com/consensys/orchestrate/src/api/service/types"
	qkmtypes "github.com/consensys/quorum-key-manager/src/stores/api/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	entitiestestdata "github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type accountsTestSuite struct {
	suite.Suite
	client            client.OrchestrateClient
	env               *IntegrationEnvironment
	defaultQKMStoreID string
}

func (s *accountsTestSuite) TestCreate() {
	ctx := s.env.ctx
	chain := entitiestestdata.FakeChain()
	chain.URLs = []string{s.env.blockchainNodeURL}

	s.T().Run("should create account successfully by querying key-manager API", func(t *testing.T) {
		txRequest := testdata.FakeCreateAccountRequest()

		ethAccRes, err := s.client.CreateAccount(ctx, txRequest)
		require.NoError(s.T(), err)

		resp, err := s.client.GetAccount(ctx, ethcommon.HexToAddress(ethAccRes.Address))
		require.NoError(s.T(), err)

		assert.Equal(s.T(), resp.Address, ethAccRes.Address)
		assert.Equal(s.T(), resp.PublicKey, ethAccRes.PublicKey)
		assert.Equal(s.T(), resp.Alias, txRequest.Alias)
		assert.Equal(s.T(), resp.StoreID, s.defaultQKMStoreID)
		assert.Equal(s.T(), resp.TenantID, "_")
	})

	s.T().Run("should fail to create account if QKM storeID does not exist", func(t *testing.T) {
		qkmStoreID := "my-personal-storeID"
		txRequest := testdata.FakeCreateAccountRequest()
		txRequest.StoreID = qkmStoreID

		_, err := s.client.CreateAccount(ctx, txRequest)
		require.Error(s.T(), err)
		// QKM StoreID does not exist
		assert.Equal(t, http.StatusFailedDependency, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should fail to create account with same alias", func(t *testing.T) {
		txRequest := testdata.FakeCreateAccountRequest()

		_, err := s.client.CreateAccount(ctx, txRequest)
		require.NoError(s.T(), err)

		_, err = s.client.CreateAccount(ctx, txRequest)
		assert.Error(s.T(), err)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())
	})

	s.T().Run("should create account successfully and trigger funding transaction", func(t *testing.T) {
		chainWithFaucet, err := s.client.RegisterChain(s.env.ctx, &api.RegisterChainRequest{
			Name: "ganache-with-faucet-accounts",
			URLs: []string{s.env.blockchainNodeURL},
			Listener: api.RegisterListenerRequest{},
		})
		require.NoError(s.T(), err)

		acc, err := account.NewAccount()
		require.NoError(s.T(), err)
		accountFaucet := entitiestestdata.FakeAccount()
		accountFaucet.Alias = "MyFaucetCreditor-accounts_" + utils.RandString(5)
		accountFaucet.Address = acc.Address

		req := testdata.FakeImportAccountRequest()
		req.PrivateKey = acc.Priv()
		req.Alias = accountFaucet.Alias
		ethAccRes, err := s.client.ImportAccount(s.env.ctx, req)
		require.NoError(s.T(), err)

		faucetRequest := testdata.FakeRegisterFaucetRequest()
		faucetRequest.Name = "faucet-integration-tests"
		faucetRequest.ChainRule = chainWithFaucet.UUID
		faucetRequest.CreditorAccount = accountFaucet.Address
		faucet, err := s.client.RegisterFaucet(s.env.ctx, faucetRequest)
		require.NoError(s.T(), err)

		accountRequest := testdata.FakeCreateAccountRequest()
		accountRequest.Chain = chainWithFaucet.Name

		require.NoError(s.T(), err)

		assert.Equal(s.T(), ethAccRes.TenantID, "_")

		err = s.client.DeleteChain(ctx, chainWithFaucet.UUID)
		assert.NoError(s.T(), err)
		err = s.client.DeleteFaucet(ctx, faucet.UUID)
		assert.NoError(s.T(), err)
	})

	s.T().Run("should fail to create account if postgres is down", func(t *testing.T) {
		txRequest := testdata.FakeCreateAccountRequest()

		err := s.env.client.Stop(ctx, postgresContainerID)
		assert.NoError(s.T(), err)

		_, err = s.client.CreateAccount(ctx, txRequest)
		assert.Error(s.T(), err)

		err = s.env.client.StartServiceAndWait(ctx, postgresContainerID, 10*time.Second)
		assert.NoError(s.T(), err)
	})
}

func (s *accountsTestSuite) TestImport() {
	ctx := s.env.ctx

	s.T().Run("should import account successfully by querying key-manager API", func(t *testing.T) {
		acc, err := account.NewAccount()
		require.NoError(s.T(), err)
		txRequest := testdata.FakeImportAccountRequest()
		txRequest.PrivateKey = acc.Priv()

		resp, err := s.client.ImportAccount(ctx, txRequest)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), resp.Alias, txRequest.Alias)
		assert.Equal(s.T(), resp.TenantID, "_")
	})

	s.T().Run("should fail to import same account twice", func(t *testing.T) {
		acc, err := account.NewAccount()
		require.NoError(s.T(), err)

		txRequest := testdata.FakeImportAccountRequest()
		txRequest.PrivateKey = acc.Priv()

		_, err = s.client.ImportAccount(ctx, txRequest)
		require.NoError(s.T(), err)

		_, err = s.client.ImportAccount(ctx, txRequest)
		require.Error(s.T(), err)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())
	})
}

func (s *accountsTestSuite) TestSearch() {
	ctx := s.env.ctx

	s.T().Run("should create account and search for it by alias successfully", func(t *testing.T) {
		txRequest := testdata.FakeCreateAccountRequest()

		ethAccRes, err := s.client.CreateAccount(ctx, txRequest)
		require.NoError(s.T(), err)
		resp, err := s.client.SearchAccounts(ctx, &entities.AccountFilters{
			Aliases: []string{txRequest.Alias},
		})
		require.NoError(s.T(), err)

		assert.Len(s.T(), resp, 1)
		assert.Equal(s.T(), resp[0].Address, ethAccRes.Address)
		assert.Equal(s.T(), resp[0].PublicKey, ethAccRes.PublicKey)
		assert.Equal(s.T(), resp[0].Alias, txRequest.Alias)
		assert.Equal(s.T(), resp[0].TenantID, "_")
	})
}

func (s *accountsTestSuite) TestUpdate() {
	ctx := s.env.ctx

	s.T().Run("should create account and update it successfully", func(t *testing.T) {
		txRequest := testdata.FakeCreateAccountRequest()

		ethAccRes, err := s.client.CreateAccount(ctx, txRequest)
		require.NoError(s.T(), err)

		txRequest2 := testdata.FakeUpdateAccountRequest()
		resp, err := s.client.UpdateAccount(ctx, ethcommon.HexToAddress(ethAccRes.Address), txRequest2)
		require.NoError(s.T(), err)

		assert.Equal(s.T(), resp.Alias, txRequest2.Alias)
		assert.Equal(s.T(), resp.Attributes, txRequest2.Attributes)
		assert.Equal(s.T(), resp.TenantID, "_")
	})
}

func (s *accountsTestSuite) TestSignMessageAndVerify() {
	ctx := s.env.ctx
	txRequest := testdata.FakeCreateAccountRequest()
	ethAccRes, err := s.client.CreateAccount(ctx, txRequest)
	require.NoError(s.T(), err)

	message := hexutil.MustDecode("0xaeff")
	var signedPayload string

	s.T().Run("should sign message successfully", func(t *testing.T) {
		signedPayload, err = s.client.SignMessage(ctx, ethcommon.HexToAddress(ethAccRes.Address), &qkmtypes.SignMessageRequest{
			Message: message,
		})
		require.NoError(s.T(), err)
		assert.NotEmpty(s.T(), signedPayload)
	})

	s.T().Run("should verify signature successfully", func(t *testing.T) {
		verifyRequest := &utilstypes.VerifyRequest{
			Data:      message,
			Signature: hexutil.MustDecode(signedPayload),
			Address:   ethcommon.HexToAddress(ethAccRes.Address),
		}
		err := s.client.VerifyMessageSignature(ctx, verifyRequest)
		assert.NoError(s.T(), err)
	})
}

func (s *accountsTestSuite) TestSignTypedData() {
	ctx := s.env.ctx

	txRequest := testdata.FakeCreateAccountRequest()
	ethAccRes, err := s.client.CreateAccount(ctx, txRequest)
	require.NoError(s.T(), err)

	typedDataRequest := qkm.FakeSignTypedDataRequest()
	var signature string

	s.T().Run("should sign typed data successfully", func(t *testing.T) {
		signature, err = s.client.SignTypedData(ctx, ethcommon.HexToAddress(ethAccRes.Address), &qkmtypes.SignTypedDataRequest{
			DomainSeparator: typedDataRequest.DomainSeparator,
			Types:           typedDataRequest.Types,
			Message:         typedDataRequest.Message,
			MessageType:     typedDataRequest.MessageType,
		})

		require.NoError(s.T(), err)
		assert.NotEmpty(s.T(), signature)
	})

	s.T().Run("should verify typed data signature successfully", func(t *testing.T) {
		err := s.client.VerifyTypedDataSignature(ctx, &utilstypes.VerifyTypedDataRequest{
			TypedData: *typedDataRequest,
			Signature: hexutil.MustDecode(signature),
			Address:   ethcommon.HexToAddress(ethAccRes.Address),
		})
		assert.NoError(s.T(), err)
	})
}
