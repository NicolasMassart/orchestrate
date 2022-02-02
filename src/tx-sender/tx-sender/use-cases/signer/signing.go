package signer

import (
	"crypto/ecdsa"
	"math/big"

	quorumtypes "github.com/consensys/quorum/core/types"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func signTransaction(tx *types.Transaction, privKey *ecdsa.PrivateKey, signer types.Signer) ([]byte, error) {
	h := signer.Hash(tx)
	decodedSignature, err := crypto.Sign(h[:], privKey)
	if err != nil {
		return nil, errors.CryptoOperationError(err.Error())
	}

	return decodedSignature, nil
}

func signQuorumPrivateTransaction(tx *quorumtypes.Transaction, privKey *ecdsa.PrivateKey, signer quorumtypes.Signer) ([]byte, error) {
	h := signer.Hash(tx)
	decodedSignature, err := crypto.Sign(h[:], privKey)
	if err != nil {
		return nil, errors.CryptoOperationError(err.Error())
	}

	return decodedSignature, nil
}

func signEEATransaction(tx *types.Transaction, privateArgs *privateETHTransactionParams, chainID *big.Int, privKey *ecdsa.PrivateKey) ([]byte, error) {
	hash, err := eeaTransactionPayload(tx, privateArgs, chainID)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(hash, privKey)
	if err != nil {
		return nil, errors.CryptoOperationError("failed to sign eea transaction").AppendReason(err.Error())
	}

	return signature, err
}
