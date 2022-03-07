package txlistener

import (
	"context"
	"fmt"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client/mock"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	testdata2 "github.com/consensys/orchestrate/pkg/types/ethereum/testdata"
	"github.com/consensys/orchestrate/src/api/service/types"
	storemocks "github.com/consensys/orchestrate/src/chain-listener/store/mocks"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterDeployedContract_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apiClient := mock.NewMockOrchestrateClient(ctrl)
	ethClient := mock2.NewMockMultiClient(ctrl)
	chainState := storemocks.NewMockChain(ctrl)
	logger := log.NewLogger()

	proxyURL := "http://api"
	apiClient.EXPECT().ChainProxyURL(gomock.Any()).AnyTimes().Return(proxyURL)

	expectedErr := fmt.Errorf("expected_err")

	chain := testdata.FakeChain()
	usecase := RegisterDeployedContractUseCase(apiClient, ethClient, chainState, logger)

	t.Run("should register deployed contract successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		job.Receipt.ContractAddress = testdata.FakeAddress().String()

		codeAt := []byte(testdata.FakeHash())
		ethClient.EXPECT().CodeAt(gomock.Any(), proxyURL, ethcommon.HexToAddress(job.Receipt.ContractAddress), gomock.Any()).
			Return(codeAt, nil)
		chainState.EXPECT().Get(gomock.Any(), job.ChainUUID).Return(chain, nil)
		apiClient.EXPECT().SetContractAddressCodeHash(gomock.Any(), job.Receipt.ContractAddress, chain.ChainID.String(), &types.SetContractCodeHashRequest{
			CodeHash: crypto.Keccak256Hash(codeAt).Bytes(),
		}).Return(nil)

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
	})

	t.Run("should register private deployed contract successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		job.Receipt.ContractAddress = testdata.FakeAddress().String()
		job.Receipt.PrivacyGroupId = "PrivacyGroupID"

		codeAt := []byte(testdata.FakeHash())
		ethClient.EXPECT().PrivCodeAt(gomock.Any(), proxyURL, ethcommon.HexToAddress(job.Receipt.ContractAddress), job.Receipt.PrivacyGroupId,
			gomock.Any()).Return(codeAt, nil)
		chainState.EXPECT().Get(gomock.Any(), job.ChainUUID).Return(chain, nil)
		apiClient.EXPECT().SetContractAddressCodeHash(gomock.Any(), job.Receipt.ContractAddress, chain.ChainID.String(), &types.SetContractCodeHashRequest{
			CodeHash: crypto.Keccak256Hash(codeAt).Bytes(),
		}).Return(nil)

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
	})

	t.Run("should fail if set contract address code hash fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		job.Receipt.ContractAddress = testdata.FakeAddress().String()

		codeAt := []byte(testdata.FakeHash())
		ethClient.EXPECT().CodeAt(gomock.Any(), proxyURL, ethcommon.HexToAddress(job.Receipt.ContractAddress), gomock.Any()).
			Return(codeAt, nil)
		chainState.EXPECT().Get(gomock.Any(), job.ChainUUID).Return(chain, nil)
		apiClient.EXPECT().SetContractAddressCodeHash(gomock.Any(), job.Receipt.ContractAddress, chain.ChainID.String(), &types.SetContractCodeHashRequest{
			CodeHash: crypto.Keccak256Hash(codeAt).Bytes(),
		}).Return(expectedErr)

		err := usecase.Execute(ctx, job)
		assert.Error(t, err)
		assert.True(t, errors.IsDependencyFailureError(err))
	})

	t.Run("should fail if get code hash fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Receipt = testdata2.FakeReceipt()
		job.Receipt.ContractAddress = testdata.FakeAddress().String()

		ethClient.EXPECT().CodeAt(gomock.Any(), proxyURL, ethcommon.HexToAddress(job.Receipt.ContractAddress), gomock.Any()).
			Return(nil, expectedErr)

		err := usecase.Execute(ctx, job)
		assert.Error(t, err)
		assert.True(t, errors.IsDependencyFailureError(err))
	})
}
