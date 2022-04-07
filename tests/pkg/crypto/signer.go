package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

func SignTransaction(privKey *ecdsa.PrivateKey, transaction *entities.ETHTransaction, chainID *big.Int) (signedRaw hexutil.Bytes, txHash *ethcommon.Hash, err error) {
	tx := transaction.ToETHTransaction(chainID)

	signer := types.NewEIP155Signer(chainID)
	decodedSignature, err := signTransaction(tx, privKey, signer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign Ethereum transaction. %v", err.Error())
	}

	signedTx, err := tx.WithSignature(signer, decodedSignature)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set transaction signature. %v", err.Error())
	}

	signedRawB, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to RLP encode signed transaction. %v", err.Error())
	}

	return signedRawB, utils.ToPtr(signedTx.Hash()).(*ethcommon.Hash), nil
}

func CreateSecp256k1(importedPrivKey []byte) (*ecdsa.PrivateKey, error) {
	var ecdsaKey *ecdsa.PrivateKey
	var err error
	if importedPrivKey != nil {
		ecdsaKey, err = crypto.ToECDSA(importedPrivKey)
		if err != nil {
			return nil, err
		}
	} else {
		ecdsaKey, err = crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
	}

	return ecdsaKey, nil
}

func signTransaction(tx *types.Transaction, privKey *ecdsa.PrivateKey, signer types.Signer) ([]byte, error) {
	h := signer.Hash(tx)
	decodedSignature, err := crypto.Sign(h[:], privKey)
	if err != nil {
		return nil, err
	}

	return decodedSignature, nil
}
