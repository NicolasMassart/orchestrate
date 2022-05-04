package api

import (
	"math/big"
	"reflect"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-playground/validator/v10"
)

var (
	validate      *validator.Validate
	StringPtrType = reflect.TypeOf(new(string))
	StringType    = reflect.TypeOf("")
)

func isHex(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		return utils.IsHexString(fl.Field().String())
	}

	return true
}

func isHexAddress(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		return ethcommon.IsHexAddress(fl.Field().String())
	}

	return true
}

func isBig(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		_, ok := new(big.Int).SetString(fl.Field().String(), 10)
		return ok
	}

	return true
}

func isTransactionType(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch entities.TransactionType(fl.Field().String()) {
		case entities.LegacyTxType, entities.DynamicFeeTxType:
			return true
		default:
			return false
		}
	}

	return true
}

func isPrivacyFlag(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch entities.PrivacyFlag(fl.Field().Int()) {
		case entities.PrivacyFlagSP, entities.PrivacyFlagPP, entities.PrivacyFlagMPP, entities.PrivacyFlagPSV:
			return true
		default:
			return false
		}
	}

	return true
}

func isHash(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		return IsHash(fl.Field().String())
	}

	return true
}

func IsHash(input string) bool {
	hash, err := hexutil.Decode(input)
	if err != nil || len(hash) != ethcommon.HashLength {
		return false
	}

	return true
}

func isDuration(fl validator.FieldLevel) bool {
	_, err := convDuration(fl)
	return err == nil
}

func minDuration(fl validator.FieldLevel) bool {
	min, err := time.ParseDuration(fl.Param())
	if err != nil {
		return false
	}

	v, err := convDuration(fl)
	if err != nil {
		return false
	}

	if v != 0 && v.Milliseconds() < min.Milliseconds() {
		return false
	}

	return true
}

func convDuration(fl validator.FieldLevel) (time.Duration, error) {
	switch fl.Field().Type() {
	case StringPtrType:
		val := fl.Field().Interface().(*string)
		if val != nil {
			return time.ParseDuration(*val)
		}
		return time.Duration(0), nil
	case StringType:
		if fl.Field().String() != "" {
			return time.ParseDuration(fl.Field().String())
		}
		return time.Duration(0), nil
	default:
		return time.Duration(0), nil
	}
}

func isPrivateTxManagerType(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch fl.Field().String() {
		case string(entities.GoQuorumChainType), string(entities.EEAChainType):
			return true
		default:
			return false
		}
	}

	return true
}

func isPriority(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch fl.Field().String() {
		case entities.PriorityVeryLow, entities.PriorityLow, entities.PriorityMedium, entities.PriorityHigh, entities.PriorityVeryHigh:
			return true
		default:
			return false
		}
	}

	return true
}

func isJobType(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch entities.JobType(fl.Field().String()) {
		case
			entities.EthereumTransaction,
			entities.EthereumRawTransaction,
			entities.EEAPrivateTransaction,
			entities.EEAMarkingTransaction,
			entities.GoQuorumPrivateTransaction,
			entities.GoQuorumMarkingTransaction:
			return true
		default:
			return false
		}
	}

	return true
}

func isJobStatus(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch entities.JobStatus(fl.Field().String()) {
		case
			entities.StatusCreated,
			entities.StatusStarted,
			entities.StatusPending,
			entities.StatusRecovering,
			entities.StatusWarning,
			entities.StatusMined,
			entities.StatusFailed,
			entities.StatusStored,
			entities.StatusResending:
			return true
		default:
			return false
		}
	}

	return true
}

func isGasIncrementLevel(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch fl.Field().String() {
		case entities.GasIncrementVeryLow.String(), entities.GasIncrementLow.String(), entities.GasIncrementMedium.String(), entities.GasIncrementHigh.String(), entities.GasIncrementVeryHigh.String():
			return true
		default:
			return false
		}
	}

	return true
}

func isKeyType(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch fl.Field().String() {
		case entities.Secp256k1:
			return true
		default:
			return false
		}
	}

	return true
}

func isEventStreamStatus(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		switch fl.Field().String() {
		case string(entities.EventStreamStatusLive), string(entities.EventStreamStatusSuspend):
			return true
		default:
			return false
		}
	}

	return true
}

func init() {
	if validate != nil {
		return
	}

	validate = validator.New()
	_ = validate.RegisterValidation("isHex", isHex)
	_ = validate.RegisterValidation("isHexAddress", isHexAddress)
	_ = validate.RegisterValidation("isBig", isBig)
	_ = validate.RegisterValidation("isHash", isHash)
	_ = validate.RegisterValidation("isDuration", isDuration)
	_ = validate.RegisterValidation("minDuration", minDuration)
	_ = validate.RegisterValidation("isPrivateTxManagerType", isPrivateTxManagerType)
	_ = validate.RegisterValidation("isPriority", isPriority)
	_ = validate.RegisterValidation("isJobType", isJobType)
	_ = validate.RegisterValidation("isJobStatus", isJobStatus)
	_ = validate.RegisterValidation("isGasIncrementLevel", isGasIncrementLevel)
	_ = validate.RegisterValidation("isKeyType", isKeyType)
	_ = validate.RegisterValidation("isTransactionType", isTransactionType)
	_ = validate.RegisterValidation("isPrivacyFlag", isPrivacyFlag)
	_ = validate.RegisterValidation("isEventStreamStatus", isEventStreamStatus)
}

func GetValidator() *validator.Validate {
	return validate
}
