// +build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	pkgutils "github.com/consensys/orchestrate/tests/pkg/utils"
	"github.com/consensys/quorum-key-manager/pkg/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type faucetTestSuite struct {
	suite.Suite
	env    *Environment
	ctx    context.Context
	cancel context.CancelFunc
}

func TestFaucets(t *testing.T) {
	s := new(faucetTestSuite)
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

func (s *faucetTestSuite) SetupSuite() {
	err := s.env.Start()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test suite has completed")
}

func (s *faucetTestSuite) TearDownSuite() {
	err := s.env.Stop()
	require.NoError(s.T(), err)
	s.env.Logger.Info("setup test teardown has completed")
}

func (s *faucetTestSuite) TestFaucets_SuccessfulUserStories() {
	faucetAccRes, err := pkgutils.ImportOrFetchAccount(s.ctx, s.env.Client, s.env.TestData.Nodes.Geth[0].FundedPublicKeys[0], &types.ImportAccountRequest{
		Alias:      "faucet-acc-" + common.RandString(5),
		PrivateKey: s.env.TestData.Nodes.Geth[0].FundedPrivateKeys[0],
	})
	require.NoError(s.T(), err)

	gethChain, _, err := s.env.createChainWithStream("chain-geth-"+common.RandString(5), s.env.TestData.Nodes.Geth[0].URLs, "")
	require.NoError(s.T(), err)

	s.T().Run("as a user I want to create faucet and be funded by it on new accounts", func(t *testing.T) {
		faucetRes, err := s.env.Client.RegisterFaucet(s.ctx, &types.RegisterFaucetRequest{
			Name:            "faucet-" + common.RandString(5),
			ChainRule:       gethChain.UUID,
			CreditorAccount: ethcommon.HexToAddress(faucetAccRes.Address),
			Cooldown:        "1s",
			MaxBalance:      *utils.HexToBigInt("0x38d7ea4c68000"),
			Amount:          *utils.HexToBigInt("0x38d7ea4c68000"),
		})
		require.NoError(t, err)
		defer func() {
			err := s.env.Client.DeleteFaucet(s.ctx, faucetRes.UUID)
			require.NoError(t, err)
		}()
	
		accRes, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{
			Chain: gethChain.Name,
		})
		require.NoError(t, err)
	
		balance, err := pkgutils.WaitForBalance(s.ctx, s.env.EthClient, s.env.Client.ChainProxyURL(gethChain.UUID), accRes.Address)
		require.NoError(t, err)
		assert.Equal(t, faucetRes.Amount, hexutil.EncodeBig(balance))
	})

	s.T().Run("as a user I want to create faucet and be funded when sending transactions", func(t *testing.T) {
		faucetRes, err := s.env.Client.RegisterFaucet(s.ctx, &types.RegisterFaucetRequest{
			Name:            "faucet-" + common.RandString(5),
			ChainRule:       gethChain.UUID,
			CreditorAccount: ethcommon.HexToAddress(faucetAccRes.Address),
			Cooldown:        "1s",
			MaxBalance:      *utils.HexToBigInt("0x38d7ea4c68000"),
			Amount:          *utils.HexToBigInt("0x38d7ea4c68000"),
		})
		require.NoError(t, err)
		defer func() {
			err := s.env.Client.DeleteFaucet(s.ctx, faucetRes.UUID)
			require.NoError(t, err)
		}()
	
		accRes, err := s.env.Client.CreateAccount(s.ctx, &types.CreateAccountRequest{})
		require.NoError(t, err)
	
		txReq := &types.TransferRequest{
			ChainName: gethChain.Name,
			Params: types.TransferParams{
				From:  ethcommon.HexToAddress(accRes.Address),
				To:    ethcommon.HexToAddress(faucetAccRes.Address),
				Value: utils.HexToBigInt("0x16345785D8A0000"),
			},
		}
		_, err = s.env.Client.SendTransferTransaction(s.ctx, txReq)
		require.NoError(t, err)
	
		balance, err := pkgutils.WaitForBalance(s.ctx, s.env.EthClient, s.env.Client.ChainProxyURL(gethChain.UUID), accRes.Address)
		require.NoError(t, err)
		assert.Equal(t, faucetRes.Amount, hexutil.EncodeBig(balance))
	})
}

func (s *faucetTestSuite) TestFaucets_FailureScenarios() {
}
