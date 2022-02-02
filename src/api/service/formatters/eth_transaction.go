package formatters

import (
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func FormatETHTransactionResponse(tx *entities.ETHTransaction) *types.ETHTransactionResponse {
	res := &types.ETHTransactionResponse{
		Nonce:           tx.Nonce,
		TransactionType: tx.TransactionType.String(),
		Gas:             tx.Gas,
		PrivateFrom:     tx.PrivateFrom,
		PrivateFor:      tx.PrivateFor,
		MandatoryFor:    tx.MandatoryFor,
		PrivacyGroupID:  tx.PrivacyGroupID,
		PrivacyFlag:     int(tx.PrivacyFlag),
		CreatedAt:       tx.CreatedAt,
		UpdatedAt:       tx.UpdatedAt,
	}

	if tx.Hash != nil {
		res.Hash = tx.Hash.String()
	}

	if tx.From != nil {
		res.From = tx.From.String()
	}

	if tx.To != nil {
		res.To = tx.To.String()
	}

	if tx.Value != nil {
		res.Value = tx.Value.String()
	}

	if tx.GasPrice != nil {
		res.GasPrice = tx.GasPrice.String()
	}

	if tx.GasFeeCap != nil {
		res.GasFeeCap = tx.GasFeeCap.String()
	}

	if tx.GasTipCap != nil {
		res.GasTipCap = tx.GasTipCap.String()
	}

	if len(tx.Data) != 0 {
		res.Data = tx.Data.String()
	}

	if len(tx.Raw) != 0 {
		res.Raw = tx.Raw.String()
	}

	if len(tx.EnclaveKey) != 0 {
		res.EnclaveKey = tx.EnclaveKey.String()
	}

	res.AccessList = []types.AccessTupleResponse{}
	for _, accessListItem := range tx.AccessList {
		elem := types.AccessTupleResponse{
			Address:     accessListItem.Address.String(),
			StorageKeys: []string{},
		}
		for _, storeKey := range accessListItem.StorageKeys {
			elem.StorageKeys = append(elem.StorageKeys, storeKey.String())
		}
		res.AccessList = append(res.AccessList, elem)
	}

	return res
}

func ETHTransactionResponseToEntity(tx *types.ETHTransactionResponse) *entities.ETHTransaction {
	res := &entities.ETHTransaction{
		Nonce:           tx.Nonce,
		TransactionType: entities.TransactionType(tx.TransactionType),
		Gas:             tx.Gas,
		PrivateFrom:     tx.PrivateFrom,
		PrivateFor:      tx.PrivateFor,
		MandatoryFor:    tx.MandatoryFor,
		PrivacyGroupID:  tx.PrivacyGroupID,
		PrivacyFlag:     entities.PrivacyFlag(tx.PrivacyFlag),
		CreatedAt:       tx.CreatedAt,
		UpdatedAt:       tx.UpdatedAt,
	}

	if tx.Hash != "" {
		res.Hash = utils.ToPtr(ethcommon.HexToHash(tx.Hash)).(*ethcommon.Hash)
	}

	if tx.From != "" {
		res.From = utils.ToPtr(ethcommon.HexToAddress(tx.From)).(*ethcommon.Address)
	}

	if tx.To != "" {
		res.To = utils.ToPtr(ethcommon.HexToAddress(tx.To)).(*ethcommon.Address)
	}

	if tx.Value != "" {
		res.Value = utils.StringBigIntToHex(tx.Value)
	}

	if tx.GasPrice != "" {
		res.GasPrice = utils.StringBigIntToHex(tx.GasPrice)
	}

	if tx.GasFeeCap != "" {
		res.GasFeeCap = utils.StringBigIntToHex(tx.GasFeeCap)
	}

	if tx.GasTipCap != "" {
		res.GasTipCap = utils.StringBigIntToHex(tx.GasTipCap)
	}

	if tx.Data != "" {
		res.Data = utils.StringToHexBytes(tx.Data)
	}

	if tx.Raw != "" {
		res.Raw = utils.StringToHexBytes(tx.Raw)
	}

	if tx.EnclaveKey != "" {
		res.EnclaveKey = utils.StringToHexBytes(tx.EnclaveKey)
	}

	res.AccessList = ethtypes.AccessList{}
	for _, accessListItem := range tx.AccessList {
		elem := ethtypes.AccessTuple{
			Address:     ethcommon.HexToAddress(accessListItem.Address),
			StorageKeys: []ethcommon.Hash{},
		}
		for _, storeKey := range accessListItem.StorageKeys {
			elem.StorageKeys = append(elem.StorageKeys, ethcommon.HexToHash(storeKey))
		}
		res.AccessList = append(res.AccessList, elem)
	}

	return res
}
