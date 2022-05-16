// +build unit

package sender

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/sdk/mock"
	testdata2 "github.com/consensys/orchestrate/src/api/service/types/testdata"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/entities/testdata"
	mock2 "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/mocks"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSendGoQuorumPrivate_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ec := mock2.NewMockQuorumTransactionSender(ctrl)
	crafter :=  mocks.NewMockCraftTransactionUseCase(ctrl)
	msgAPI := mock.NewMockMessengerAPI(ctrl)
	chainRegistryURL := "chainRegistryURL:8081"
	ctx := context.Background()

	usecase := NewSendGoQuorumPrivateTxUseCase(ec, crafter, msgAPI, chainRegistryURL)

	t.Run("should execute use case successfully", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.PrivateFrom = "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=" 
		job.Transaction.Data = hexutil.MustDecode("0xfe378324abcde723")
		enclaveKey, _ := base64.StdEncoding.DecodeString("ZW5jbGF2ZUtleQ==")

		proxyURL := client.GetProxyTesseraURL(chainRegistryURL, job.ChainUUID)
		
		crafter.EXPECT().Execute(gomock.Any(), job)
		ec.EXPECT().StoreRaw(gomock.Any(), proxyURL, job.Transaction.Data, job.Transaction.PrivateFrom).Return(enclaveKey, nil)

		msgAPI.EXPECT().JobUpdateMessage(gomock.Any(), 
			testdata2.SentJobMessageRequestMatcher(job.UUID, entities.StatusStored, nil), gomock.Any()).Return(nil)

		err := usecase.Execute(ctx, job)
		assert.NoError(t, err)
		assert.Equal(t, job.Transaction.EnclaveKey.String(), hexutil.Encode(enclaveKey))
	})
	
	t.Run("should fail with same error executing use case if storeRaw fails", func(t *testing.T) {
		job := testdata.FakeJob()
		job.Transaction.PrivateFrom = "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=" 
		job.Transaction.Data = hexutil.MustDecode("0xfe378324abcde723")
		enclaveKey, _ := base64.StdEncoding.DecodeString("ZW5jbGF2ZUtleQ==")

		proxyURL := client.GetProxyTesseraURL(chainRegistryURL, job.ChainUUID)
		crafter.EXPECT().Execute(gomock.Any(), job)
		expectedErr := errors.InternalError("internal_err")
		ec.EXPECT().StoreRaw(gomock.Any(), proxyURL, job.Transaction.Data, job.Transaction.PrivateFrom).Return(enclaveKey, expectedErr)

		err := usecase.Execute(ctx, job)
		assert.Equal(t, err, expectedErr)
	})
}
