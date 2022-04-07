// +build e2e

package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/api/service/types/testdata"
	testdata2 "github.com/consensys/orchestrate/src/entities/testdata"
	"github.com/consensys/orchestrate/src/infra/quorum-key-manager/testutils"
	"github.com/consensys/quorum-key-manager/pkg/common"
	types2 "github.com/consensys/quorum-key-manager/src/stores/api/types"
	types3 "github.com/consensys/quorum-key-manager/src/utils/api/types"
	"github.com/consensys/quorum/common/hexutil"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type accountTestSuite struct {
	suite.Suite
	env    *Environment
	ctx    context.Context
	cancel context.CancelFunc
}

func TestAccounts(t *testing.T) {
	s := new(accountTestSuite)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	defer s.cancel()

	var err error
	s.env, err = NewEnvironment(s.ctx, s.cancel)
	require.NoError(t, err)

	sig := common.NewSignalListener(func(signal os.Signal) {
		s.env.Logger.Error("interrupt signal was caught")
		s.cancel()
		t.FailNow()
	})

	defer sig.Close()

	suite.Run(t, s)
}

func (s *accountTestSuite) TestAccounts_SuccessfulUserStories() {
	s.T().Run("as a user I want to create account with alias, attr and alternative store then update it", func(t *testing.T) {
		fakeReq := testdata.FakeCreateAccountRequest()
		fakeReq.StoreID = s.env.TestData.QKMStoreID
		res, err := s.env.Client.CreateAccount(s.ctx, fakeReq)
		require.NoError(t, err)
		assert.Equal(t, s.env.UserInfo.TenantID, res.TenantID)
		assert.Equal(t, fakeReq.Alias, res.Alias)
		assert.Equal(t, fakeReq.Attributes, res.Attributes)
		assert.Equal(t, s.env.TestData.QKMStoreID, res.StoreID)

		fakeReq2 := testdata.FakeUpdateAccountRequest()
		res2, err := s.env.Client.UpdateAccount(s.ctx, ethcommon.HexToAddress(res.Address), fakeReq2)
		require.NoError(t, err)
		assert.Equal(t, fakeReq2.Alias, res2.Alias)
		assert.Equal(t, fakeReq2.Attributes, res2.Attributes)
	})

	s.T().Run("as a user I want to create account, sign message and verify corresponding signature", func(t *testing.T) {
		res, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(t, err)

		assert.Equal(t, s.env.UserInfo.TenantID, res.TenantID)

		signMsgReq := &types2.SignMessageRequest{
			Message: hexutil.MustDecode("0xfe"),
		}

		signature, err := s.env.Client.SignMessage(s.ctx, ethcommon.HexToAddress(res.Address), signMsgReq)
		require.NoError(t, err)
		assert.NotEmpty(t, signature)

		err = s.env.Client.VerifyMessageSignature(s.ctx, &types3.VerifyRequest{
			Data:      signMsgReq.Message,
			Address:   ethcommon.HexToAddress(res.Address),
			Signature: hexutil.MustDecode(signature),
		})

		require.NoError(t, err)
	})

	s.T().Run("as a user I want to import an account, sign typed data and verify corresponding signature", func(t *testing.T) {
		fakeReq := testdata.FakeImportAccountRequest()
		privKey, _ := crypto.GenerateKey()
		fakeReq.PrivateKey = privKey.D.Bytes()
		res, err := s.env.Client.ImportAccount(s.ctx, fakeReq)
		require.NoError(t, err)
		assert.Equal(t, s.env.UserInfo.TenantID, res.TenantID)

		fakeReq2 := testutils.FakeSignTypedDataRequest()
		signature, err := s.env.Client.SignTypedData(s.ctx, ethcommon.HexToAddress(res.Address), fakeReq2)
		require.NoError(t, err)
		assert.NotEmpty(t, signature)

		err = s.env.Client.VerifyTypedDataSignature(s.ctx, &types3.VerifyTypedDataRequest{
			TypedData: *fakeReq2,
			Address:   ethcommon.HexToAddress(res.Address),
			Signature: hexutil.MustDecode(signature),
		})

		require.NoError(t, err)
	})
}

func (s *accountTestSuite) TestAccounts_FailureScenarios() {
	s.T().Run("when an user creates two accounts with same alias if should fail with expected error", func(t *testing.T) {
		fakeReq := testdata.FakeCreateAccountRequest()
		_, err := s.env.Client.CreateAccount(s.ctx, fakeReq)
		require.NoError(t, err)
		
		_, err = s.env.Client.CreateAccount(s.ctx, fakeReq)
		require.Error(t, err)
		assert.Equal(t, http.StatusConflict, err.(*client.HTTPErr).Code())
	})
	
	s.T().Run("when an user creates an account with a not existing QKM StoreID if should fail with expected error", func(t *testing.T) {
		fakeReq := testdata.FakeCreateAccountRequest()
		fakeReq.StoreID = "notExisting"
		_, err := s.env.Client.CreateAccount(s.ctx, fakeReq)
		require.Error(t, err)
		assert.Equal(t, http.StatusFailedDependency, err.(*client.HTTPErr).Code())
	})
	
	s.T().Run("when an user wants to sign with a not existing account if should fail with expected error", func(t *testing.T) {
		signMsgReq := &types2.SignMessageRequest{
			Message: hexutil.MustDecode("0xfe"),
		}

		notExistingAddr := testdata2.FakeAddress()
		_, err := s.env.Client.SignMessage(s.ctx, *notExistingAddr, signMsgReq)
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*client.HTTPErr).Code())
	})
}
