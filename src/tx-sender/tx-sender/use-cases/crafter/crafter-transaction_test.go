// +build unit

package crafter

import (
	"context"
	"math/big"
	"testing"

	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/nonce/mocks"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCrafterTransaction_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	ec := mock2.NewMockMultiClient(ctrl)
	nm := mocks.NewMockManager(ctrl)
	chainRegistryURL := "http://chain-registry:8081"

	nextBaseFee, _ := new(big.Int).SetString("1000000000", 10)
	mediumPriority, _ := new(big.Int).SetString(mediumPriorityString, 10)

	usecase := NewCraftTransactionUseCase(ec, chainRegistryURL, nm)

	t.Run("should execute use case for LegacyTx successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.Nonce = nil
		job.Transaction.GasPrice = nil
		job.Transaction.Gas = nil
		job.Transaction.TransactionType = entities.LegacyTxType

		proxyURL := client.GetProxyURL(chainRegistryURL, job.ChainUUID)
		expectedGasPrice, _ := new(big.Int).SetString("1000", 10)
		ec.EXPECT().SuggestGasPrice(gomock.Any(), proxyURL).Return(expectedGasPrice, nil)
		ec.EXPECT().EstimateGas(gomock.Any(), proxyURL, gomock.Any()).Return(uint64(1000), nil)
		nm.EXPECT().GetNonce(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.Equal(t, expectedGasPrice.String(), job.Transaction.GasPrice.ToInt().String())
		assert.Equal(t, uint64(1000), *job.Transaction.Gas)
		assert.Equal(t, uint64(1), *job.Transaction.Nonce)
	})

	t.Run("should execute use case for DynamicFeeTx successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.Nonce = nil
		job.Transaction.Gas = nil
		job.Transaction.GasPrice = nil

		proxyURL := client.GetProxyURL(chainRegistryURL, job.ChainUUID)
		expectedFeeHistory := testdata.FakeFeeHistory(nextBaseFee)
		ec.EXPECT().FeeHistory(gomock.Any(), proxyURL, 1, "latest").Return(expectedFeeHistory, nil)
		ec.EXPECT().EstimateGas(gomock.Any(), proxyURL, gomock.Any()).Return(uint64(1000), nil)
		nm.EXPECT().GetNonce(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
		err := usecase.Execute(ctx, job)

		expectedFeeCap := new(big.Int).Add(mediumPriority, nextBaseFee)

		assert.NoError(t, err)
		assert.Equal(t, entities.DynamicFeeTxType, job.Transaction.TransactionType)
		assert.Equal(t, expectedFeeCap.String(), job.Transaction.GasFeeCap.ToInt().String())
		assert.Equal(t, mediumPriority.String(), job.Transaction.GasTipCap.ToInt().String())
		assert.Equal(t, uint64(1000), *job.Transaction.Gas)
		assert.Equal(t, uint64(1), *job.Transaction.Nonce)
	})

	t.Run("should execute use case for OneTimeKey successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.Nonce = nil
		job.Transaction.GasPrice = nil
		job.Transaction.Gas = nil
		job.InternalData.OneTimeKey = true

		proxyURL := client.GetProxyURL(chainRegistryURL, job.ChainUUID)
		expectedGasPrice, _ := new(big.Int).SetString("1000", 10)
		ec.EXPECT().SuggestGasPrice(gomock.Any(), proxyURL).Return(expectedGasPrice, nil)
		ec.EXPECT().EstimateGas(gomock.Any(), proxyURL, gomock.Any()).Return(uint64(1000), nil)
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.Equal(t, expectedGasPrice.String(), job.Transaction.GasPrice.ToInt().String())
		assert.Equal(t, uint64(1000), *job.Transaction.Gas)
		assert.Equal(t, uint64(0), *job.Transaction.Nonce)
	})

	t.Run("should execute use case for EEA marking transaction successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Type = entities.EEAMarkingTransaction
		job.Transaction.Nonce = nil
		job.Transaction.GasPrice = nil
		job.Transaction.Gas = nil

		proxyURL := client.GetProxyURL(chainRegistryURL, job.ChainUUID)
		expectedContractAddr := ethcommon.HexToAddress("0x1")
		expectedFeeHistory := testdata.FakeFeeHistory(nextBaseFee)
		ec.EXPECT().FeeHistory(gomock.Any(), proxyURL, 1, "latest").Return(expectedFeeHistory, nil)
		ec.EXPECT().EstimateGas(gomock.Any(), proxyURL, gomock.Any()).Return(uint64(1000), nil)
		ec.EXPECT().EEAPrivPrecompiledContractAddr(gomock.Any(), proxyURL).Return(expectedContractAddr, nil)
		nm.EXPECT().GetNonce(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.Equal(t, uint64(1000), *job.Transaction.Gas)
		assert.Equal(t, uint64(1), *job.Transaction.Nonce)
		assert.Equal(t, expectedContractAddr.String(), job.Transaction.To.String())
	})

	t.Run("should execute use case for EEA private transaction successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Type = entities.EEAPrivateTransaction
		job.Transaction.Nonce = nil
		job.Transaction.GasPrice = nil
		job.Transaction.Gas = nil

		nm.EXPECT().GetNonce(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
		err := usecase.Execute(ctx, job)

		assert.NoError(t, err)
		assert.Empty(t, job.Transaction.GasPrice)
		assert.Empty(t, job.Transaction.Gas)
		assert.Equal(t, uint64(1), *job.Transaction.Nonce)
	})

	t.Run("should execute use case for child job for DynamicTx successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.GasTipCap = (*hexutil.Big)(hexutil.MustDecodeBig("0x30D3F"))
		job.Transaction.TransactionType = entities.DynamicFeeTxType
		job.InternalData.ParentJobUUID = job.UUID

		proxyURL := client.GetProxyURL(chainRegistryURL, job.ChainUUID)
		expectedFeeHistory := testdata.FakeFeeHistory(nextBaseFee)
		ec.EXPECT().FeeHistory(gomock.Any(), proxyURL, 1, "latest").Return(expectedFeeHistory, nil)

		err := usecase.Execute(ctx, job)
		expectedFeeCap := new(big.Int).Add(job.Transaction.GasTipCap.ToInt(), nextBaseFee)

		assert.NoError(t, err)
		assert.Equal(t, expectedFeeCap.String(), job.Transaction.GasFeeCap.ToInt().String())
	})
}
