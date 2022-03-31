package signer

import (
	"context"
	"math/big"

	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	qkmtypes "github.com/consensys/quorum-key-manager/src/stores/api/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/crypto"

	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/quorum-key-manager/pkg/client"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

const signEEATransactionComponent = "use-cases.sign-eea-transaction"

type signEEAPrivateTransactionUseCase struct {
	keyManagerClient client.KeyManagerClient
	logger           *log.Logger
}

func NewSignEEAPrivateTransactionUseCase(keyManagerClient client.KeyManagerClient) usecases.SignEEATransactionUseCase {
	return &signEEAPrivateTransactionUseCase{
		keyManagerClient: keyManagerClient,
		logger:           log.NewLogger().SetComponent(signEEATransactionComponent),
	}
}

func (uc *signEEAPrivateTransactionUseCase) Execute(ctx context.Context, job *entities.Job) (signedRaw hexutil.Bytes, txHash *ethcommon.Hash, err error) {
	logger := uc.logger.WithContext(ctx).WithField("one_time_key", job.InternalData.OneTimeKey)

	transaction := job.Transaction.ToETHTransaction(job.InternalData.ChainID)
	privateArgs := &privateETHTransactionParams{
		PrivateFrom:    job.Transaction.PrivateFrom,
		PrivateFor:     job.Transaction.PrivateFor,
		PrivacyGroupID: job.Transaction.PrivacyGroupID,
		PrivateTxType:  entities.PrivateTxTypeRestricted.String(),
	}

	if job.InternalData.OneTimeKey {
		signedRaw, err = uc.signWithOneTimeKey(ctx, transaction, privateArgs, job.InternalData.ChainID)
	} else {
		signedRaw, err = uc.signWithAccount(ctx, job, privateArgs, transaction, job.InternalData.ChainID)
	}
	if err != nil {
		return nil, nil, errors.FromError(err).ExtendComponent(signEEATransactionComponent)
	}

	logger.Debug("eea transaction signed successfully")

	// transaction hash of EEA transactions cannot be computed
	return signedRaw, nil, nil
}

func (uc *signEEAPrivateTransactionUseCase) signWithOneTimeKey(ctx context.Context, transaction *types.Transaction,
	privateArgs *privateETHTransactionParams, chainID *big.Int) (hexutil.Bytes, error) {
	logger := uc.logger.WithContext(ctx)
	privKey, err := crypto.GenerateKey()
	if err != nil {
		errMessage := "failed to generate ethereum private key"
		logger.WithError(err).Error(errMessage)
		return nil, errors.CryptoOperationError(errMessage)
	}

	decodedSignature, err := signEEATransaction(transaction, privateArgs, chainID, privKey)
	if err != nil {
		logger.WithError(err).Error("failed to sign EEA transaction")
		return nil, err
	}

	signedRaw, err := uc.getSignedRawEEATransaction(ctx, transaction, privateArgs, decodedSignature, chainID)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(signEEATransactionComponent)
	}

	logger.Debug("eea transaction signed successfully")

	return signedRaw, nil
}

func (uc *signEEAPrivateTransactionUseCase) signWithAccount(ctx context.Context, job *entities.Job,
	privateArgs *privateETHTransactionParams, tx *types.Transaction, chainID *big.Int) (hexutil.Bytes, error) {
	logger := uc.logger.WithContext(ctx)

	req := &qkmtypes.SignEEATransactionRequest{
		Nonce:          hexutil.Uint64(tx.Nonce()),
		To:             tx.To(),
		Data:           tx.Data(),
		Value:          hexutil.Big(*tx.Value()),
		GasPrice:       hexutil.Big(*tx.GasPrice()),
		GasLimit:       hexutil.Uint64(tx.Gas()),
		ChainID:        hexutil.Big(*chainID),
		PrivateFrom:    privateArgs.PrivateFrom,
		PrivateFor:     privateArgs.PrivateFor,
		PrivacyGroupID: privateArgs.PrivacyGroupID,
	}

	signedRawStr, err := uc.keyManagerClient.SignEEATransaction(ctx, job.InternalData.StoreID, job.Transaction.From.Hex(), req)

	if err != nil {
		errMsg := "failed to sign eea transaction using key manager"
		logger.WithError(err).Error(errMsg)
		return nil, errors.DependencyFailureError(errMsg).AppendReason(err.Error())
	}

	signedRaw, err := hexutil.Decode(signedRawStr)
	if err != nil {
		errMessage := "failed to decode eea raw signature"
		logger.WithError(err).Error(errMessage)
		return nil, errors.EncodingError(errMessage)
	}

	return signedRaw, nil
}

func (uc *signEEAPrivateTransactionUseCase) getSignedRawEEATransaction(ctx context.Context, transaction *types.Transaction,
	privateArgs *privateETHTransactionParams, signature []byte, chainID *big.Int) (hexutil.Bytes, error) {
	logger := uc.logger.WithContext(ctx)

	signedTx, err := transaction.WithSignature(types.NewEIP155Signer(chainID), signature)
	if err != nil {
		errMessage := "failed to set eea transaction signature"
		logger.WithError(err).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage)
	}
	v, r, s := signedTx.RawSignatureValues()

	privateFromEncoded, err := getEncodedPrivateFrom(privateArgs.PrivateFrom)
	if err != nil {
		return nil, err
	}

	privateRecipientEncoded, err := getEncodedPrivateRecipient(privateArgs.PrivacyGroupID, privateArgs.PrivateFor)
	if err != nil {
		return nil, err
	}

	signedRaw, err := rlp.EncodeToBytes([]interface{}{
		transaction.Nonce(),
		transaction.GasPrice(),
		transaction.Gas(),
		transaction.To(),
		transaction.Value(),
		transaction.Data(),
		v,
		r,
		s,
		privateFromEncoded,
		privateRecipientEncoded,
		privateArgs.PrivateTxType,
	})
	if err != nil {
		errMessage := "failed to RLP encode signed eea transaction"
		logger.WithError(err).Error(errMessage)
		return nil, errors.CryptoOperationError(errMessage)
	}

	return signedRaw, nil
}
