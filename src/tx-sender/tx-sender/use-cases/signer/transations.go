package signer

import (
	"encoding/base64"
	"math/big"

	"github.com/consensys/orchestrate/pkg/errors"
	quorumtypes "github.com/consensys/quorum/core/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

type privateETHTransactionParams struct {
	PrivateFrom    string
	PrivateFor     []string
	PrivacyGroupID string
	PrivateTxType  string
}

func eeaTransactionPayload(tx *types.Transaction, privateArgs *privateETHTransactionParams, chainID *big.Int) ([]byte, error) {
	privateFromEncoded, err := getEncodedPrivateFrom(privateArgs.PrivateFrom)
	if err != nil {
		return nil, err
	}

	privateRecipientEncoded, err := getEncodedPrivateRecipient(privateArgs.PrivacyGroupID, privateArgs.PrivateFor)
	if err != nil {
		return nil, err
	}

	hash, err := hash([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		chainID,
		uint(0),
		uint(0),
		privateFromEncoded,
		privateRecipientEncoded,
		privateArgs.PrivateTxType,
	})
	if err != nil {
		return nil, errors.CryptoOperationError("failed to hash eea transaction").AppendReason(err.Error())
	}

	return hash.Bytes(), nil
}

func hash(object interface{}) (hash ethcommon.Hash, err error) {
	hashAlgo := sha3.NewLegacyKeccak256()
	err = rlp.Encode(hashAlgo, object)
	if err != nil {
		return ethcommon.Hash{}, err
	}
	hashAlgo.Sum(hash[:0])
	return hash, nil
}

func getEncodedPrivateFrom(privateFrom string) ([]byte, error) {
	privateFromEncoded, err := base64.StdEncoding.DecodeString(privateFrom)
	if err != nil {
		return nil, errors.EncodingError("invalid base64 value for 'privateFrom'").AppendReason(err.Error())
	}

	return privateFromEncoded, nil
}

func getEncodedPrivateRecipient(privacyGroupID string, privateFor []string) (interface{}, error) {
	var privateRecipientEncoded interface{}
	var err error
	if privacyGroupID != "" {
		privateRecipientEncoded, err = base64.StdEncoding.DecodeString(privacyGroupID)
		if err != nil {
			return nil, errors.EncodingError("invalid base64 value for 'privacyGroupId'").AppendReason(err.Error())
		}
	} else {
		var privateForByteSlice [][]byte
		for _, v := range privateFor {
			b, der := base64.StdEncoding.DecodeString(v)
			if der != nil {
				return nil, errors.EncodingError("invalid base64 value for 'privateFor'").AppendReason(der.Error())
			}
			privateForByteSlice = append(privateForByteSlice, b)
		}
		privateRecipientEncoded = privateForByteSlice
	}

	return privateRecipientEncoded, nil
}

func getQuorumPrivateTxSigner() quorumtypes.Signer {
	return quorumtypes.QuorumPrivateTxSigner{}
}
