package signer

import (
	"context"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	qkmtypes "github.com/consensys/quorum-key-manager/src/stores/api/types"
	quorumtypes "github.com/consensys/quorum/core/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/quorum-key-manager/pkg/client"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const signQuorumPrivateTransactionComponent = "use-cases.sign-quorum-private-transaction"

// signQuorumPrivateTransactionUseCase is a use case to sign a quorum private transaction
type signGoQuorumPrivateTransactionUseCase struct {
	keyManagerClient client.KeyManagerClient
	logger           *log.Logger
}

func NewSignGoQuorumPrivateTransactionUseCase(keyManagerClient client.KeyManagerClient) usecases.SignGoQuorumPrivateTransactionUseCase {
	return &signGoQuorumPrivateTransactionUseCase{
		keyManagerClient: keyManagerClient,
		logger:           log.NewLogger().SetComponent(signQuorumPrivateTransactionComponent),
	}
}

// Execute signs a quorum private transaction
func (uc *signGoQuorumPrivateTransactionUseCase) Execute(ctx context.Context, job *entities.Job) (signedRaw hexutil.Bytes, txHash *ethcommon.Hash, err error) {
	logger := uc.logger.WithContext(ctx).WithField("one_time_key", job.InternalData.OneTimeKey)

	transaction := job.Transaction.ToQuorumTransaction()
	transaction.SetPrivate()
	if job.InternalData.OneTimeKey {
		signedRaw, txHash, err = uc.signWithOneTimeKey(ctx, transaction)
	} else {
		signedRaw, txHash, err = uc.signWithAccount(ctx, job, transaction)
	}

	if err != nil {
		return nil, nil, errors.FromError(err).ExtendComponent(signQuorumPrivateTransactionComponent)
	}

	logger.WithField("tx_hash", txHash).Debug("quorum private transaction signed successfully")
	return signedRaw, txHash, nil
}

func (uc *signGoQuorumPrivateTransactionUseCase) signWithOneTimeKey(ctx context.Context, transaction *quorumtypes.Transaction) (
	signedRaw hexutil.Bytes, txHash *ethcommon.Hash, err error) {
	logger := uc.logger.WithContext(ctx)
	privKey, err := crypto.GenerateKey()
	if err != nil {
		errMessage := "failed to generate Ethereum account"
		logger.WithError(err).Error(errMessage)
		return nil, nil, errors.CryptoOperationError(errMessage)
	}

	signer := getQuorumPrivateTxSigner()
	decodedSignature, err := signQuorumPrivateTransaction(transaction, privKey, signer)
	if err != nil {
		logger.WithError(err).Error("failed to sign private transaction")
		return nil, nil, err
	}

	signedTx, err := transaction.WithSignature(signer, decodedSignature)
	if err != nil {
		errMessage := "failed to set quorum private transaction signature"
		logger.WithError(err).Error(errMessage)
		return nil, nil, errors.InvalidParameterError(errMessage).ExtendComponent(signQuorumPrivateTransactionComponent)
	}

	signedRawB, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		errMessage := "failed to RLP encode signed quorum private transaction"
		logger.WithError(err).Error(errMessage)
		return nil, nil, errors.CryptoOperationError(errMessage).ExtendComponent(signQuorumPrivateTransactionComponent)
	}

	return signedRawB, utils.ToPtr(signedTx.Hash()).(*ethcommon.Hash), nil
}

func (uc *signGoQuorumPrivateTransactionUseCase) signWithAccount(ctx context.Context, job *entities.Job, tx *quorumtypes.Transaction) (
	signedRaw hexutil.Bytes, txHash *ethcommon.Hash, err error) {
	logger := uc.logger.WithContext(ctx)

	signedRawStr, err := uc.keyManagerClient.SignQuorumPrivateTransaction(ctx, job.InternalData.StoreID, job.Transaction.From.Hex(), &qkmtypes.SignQuorumPrivateTransactionRequest{
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    hexutil.Big(*tx.Value()),
		GasPrice: hexutil.Big(*tx.GasPrice()),
		GasLimit: hexutil.Uint64(tx.Gas()),
		Data:     tx.Data(),
	})
	if err != nil {
		errMsg := "failed to sign quorum private transaction using key manager"
		logger.WithError(err).Error(errMsg)
		return nil, nil, errors.DependencyFailureError(errMsg).AppendReason(err.Error())
	}

	signedRaw, err = hexutil.Decode(signedRawStr)
	if err != nil {
		errMessage := "failed to decode quorum raw signature"
		logger.WithError(err).Error(errMessage)
		return nil, nil, errors.EncodingError(errMessage)
	}

	err = rlp.DecodeBytes(signedRaw, &tx)
	if err != nil {
		errMessage := "failed to decode quorum transaction"
		logger.WithError(err).Error(errMessage)
		return nil, nil, errors.EncodingError(errMessage)
	}

	return signedRaw, utils.ToPtr(tx.Hash()).(*ethcommon.Hash), nil
}
