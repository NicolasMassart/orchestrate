package transactions

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofrs/uuid"
)

const sendTxComponent = "use-cases.send-tx"

type sendTxUsecase struct {
	db                 store.DB
	searchChainsUC     usecases.SearchChainsUseCase
	startJobUC         usecases.StartJobUseCase
	createJobUC        usecases.CreateJobUseCase
	getTxUC            usecases.GetTxUseCase
	getFaucetCandidate usecases.GetFaucetCandidateUseCase
	logger             *log.Logger
}

func NewSendTxUseCase(
	db store.DB,
	searchChainsUC usecases.SearchChainsUseCase,
	startJobUseCase usecases.StartJobUseCase,
	createJobUC usecases.CreateJobUseCase,
	getTxUC usecases.GetTxUseCase,
	getFaucetCandidate usecases.GetFaucetCandidateUseCase,
) usecases.SendTxUseCase {
	return &sendTxUsecase{
		db:                 db,
		searchChainsUC:     searchChainsUC,
		startJobUC:         startJobUseCase,
		createJobUC:        createJobUC,
		getTxUC:            getTxUC,
		getFaucetCandidate: getFaucetCandidate,
		logger:             log.NewLogger().SetComponent(sendTxComponent),
	}
}

func (uc *sendTxUsecase) Execute(ctx context.Context, txRequest *entities.TxRequest, txData hexutil.Bytes, userInfo *multitenancy.UserInfo) (*entities.TxRequest, error) {
	ctx = log.WithFields(ctx, log.Field("idempotency-key", txRequest.IdempotencyKey))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("creating new transaction")

	// Step 1: Get chain from chain registry
	chain, err := uc.getChain(ctx, txRequest.ChainName, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 2: Generate request hash
	requestHash, err := generateRequestHash(chain.UUID, txRequest.Params)
	if err != nil {
		errMessage := "failed to generate request hash"
		logger.WithError(err).Error(errMessage)
		return nil, errors.InvalidParameterError("failed to generate request hash").ExtendComponent(sendTxComponent)
	}

	// Step 3: Insert Schedule + Job + Transaction + TxRequest atomically OR get tx request if it exists
	txRequest, err = uc.selectOrInsertTxRequest(ctx, txRequest, txData, requestHash, chain.UUID, userInfo)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 4: Start first job of the schedule if status is CREATED
	// Otherwise there was another request with same idempotency key and same reqHash
	job := txRequest.Schedule.Jobs[0]
	if job.Status == entities.StatusCreated {
		var fctJob *entities.Job
		fctJob, err = uc.startFaucetJob(ctx, job.Transaction.From, job.ScheduleUUID, chain, userInfo)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
		if fctJob != nil {
			txRequest.Schedule.Jobs = append(txRequest.Schedule.Jobs, fctJob)
		}

		if err = uc.startJobUC.Execute(ctx, job.UUID, userInfo); err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
	} else { // Load the latest Schedule status from DB
		txRequest, err = uc.getTxUC.Execute(ctx, txRequest.Schedule.UUID, userInfo)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
	}

	logger.WithField("schedule", txRequest.Schedule.UUID).Info("transaction created successfully")
	return txRequest, nil
}

func (uc *sendTxUsecase) getChain(ctx context.Context, chainName string, userInfo *multitenancy.UserInfo) (*entities.Chain, error) {
	chains, err := uc.searchChainsUC.Execute(ctx, &entities.ChainFilters{Names: []string{chainName}}, userInfo)
	if err != nil {
		return nil, err
	}

	if len(chains) == 0 {
		errMessage := fmt.Sprintf("chain '%s' does not exist", chainName)
		uc.logger.WithContext(ctx).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage)
	}

	return chains[0], nil
}

func (uc *sendTxUsecase) selectOrInsertTxRequest(
	ctx context.Context,
	txRequest *entities.TxRequest,
	txData []byte, requestHash, chainUUID string,
	userInfo *multitenancy.UserInfo,
) (*entities.TxRequest, error) {
	if txRequest.IdempotencyKey == "" {
		return uc.insertNewTxRequest(ctx, txRequest, txData, requestHash, chainUUID, userInfo.TenantID, userInfo)
	}

	curTxRequest, err := uc.db.TransactionRequest().FindOneByIdempotencyKey(ctx, txRequest.IdempotencyKey, userInfo.TenantID, userInfo.Username)
	switch {
	case errors.IsNotFoundError(err):
		return uc.insertNewTxRequest(ctx, txRequest, txData, requestHash, chainUUID, userInfo.TenantID, userInfo)
	case err != nil:
		return nil, err
	case curTxRequest != nil && curTxRequest.Hash != requestHash:
		errMessage := "transaction request with the same idempotency key and different params already exists"
		uc.logger.Error(errMessage)
		return nil, errors.AlreadyExistsError(errMessage)
	default:
		return uc.getTxUC.Execute(ctx, curTxRequest.Schedule.UUID, userInfo)
	}
}

func (uc *sendTxUsecase) insertNewTxRequest(
	ctx context.Context,
	txRequest *entities.TxRequest,
	txData []byte, requestHash, chainUUID, tenantID string,
	userInfo *multitenancy.UserInfo,
) (*entities.TxRequest, error) {
	err := uc.db.RunInTransaction(ctx, func(dbtx store.DB) error {
		txRequest.Schedule = &entities.Schedule{TenantID: tenantID, OwnerID: userInfo.Username}
		err := dbtx.Schedule().Insert(ctx, txRequest.Schedule)
		if err != nil {
			return err
		}

		_, err = dbtx.TransactionRequest().Insert(ctx, txRequest, requestHash, txRequest.Schedule.UUID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sendTxJobs, err := uc.newJobEntities(txRequest, chainUUID, txData)
	if err != nil {
		return nil, err
	}

	txRequest.Schedule.Jobs = make([]*entities.Job, len(sendTxJobs))
	var nextJobUUID string
	for idx, txJob := range sendTxJobs {
		if nextJobUUID != "" {
			txJob.UUID = nextJobUUID
		}

		if idx < len(sendTxJobs)-1 {
			nextJobUUID = uuid.Must(uuid.NewV4()).String()
			txJob.NextJobUUID = nextJobUUID
		}

		var job *entities.Job
		job, err = uc.createJobUC.Execute(ctx, txJob, userInfo)
		if err != nil {
			return nil, err
		}

		txRequest.Schedule.Jobs[idx] = job
	}

	return txRequest, nil
}

// Execute validates, creates and starts a new transaction for pre funding users account
func (uc *sendTxUsecase) startFaucetJob(ctx context.Context, account *ethcommon.Address, scheduleUUID string,
	chain *entities.Chain, userInfo *multitenancy.UserInfo) (*entities.Job, error) {
	if account == nil {
		return nil, nil
	}

	logger := uc.logger.WithContext(ctx).WithField("chain", chain.UUID)
	faucet, err := uc.getFaucetCandidate.Execute(ctx, *account, chain, userInfo)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	logger.WithField("faucet_amount", faucet.Amount).Debug("faucet: credit approved")

	txJob := &entities.Job{
		ScheduleUUID: scheduleUUID,
		ChainUUID:    chain.UUID,
		Type:         entities.EthereumTransaction,
		Labels: map[string]string{
			"faucetUUID": faucet.UUID,
		},
		InternalData: &entities.InternalData{},
		Transaction: &entities.ETHTransaction{
			From:  &faucet.CreditorAccount,
			To:    account,
			Value: &faucet.Amount,
		},
	}

	internalAdminUser := multitenancy.NewInternalAdminUser()
	internalAdminUser.TenantID = userInfo.TenantID
	fctJob, err := uc.createJobUC.Execute(ctx, txJob, internalAdminUser)
	if err != nil {
		return nil, err
	}

	err = uc.startJobUC.Execute(ctx, fctJob.UUID, internalAdminUser)
	if err != nil {
		return fctJob, err
	}

	return fctJob, nil
}

func generateRequestHash(chainUUID string, params interface{}) (string, error) {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	hash := md5.Sum([]byte(string(jsonParams) + chainUUID))
	return hex.EncodeToString(hash[:]), nil
}

func (uc *sendTxUsecase) newJobEntities(txRequest *entities.TxRequest, chainUUID string, txData []byte) ([]*entities.Job, error) {
	var jobs []*entities.Job
	switch {
	case txRequest.Params.Protocol == entities.EEAChainType:
		privTxJob := uc.newJobEntityFromTxRequest(txRequest, uc.newEthTransactionFromParams(txRequest.Params, txData, entities.LegacyTxType), entities.EEAPrivateTransaction, chainUUID)
		markingTxJob := uc.newJobEntityFromTxRequest(txRequest, &entities.ETHTransaction{}, entities.EEAMarkingTransaction, chainUUID)
		markingTxJob.InternalData.OneTimeKey = true
		jobs = append(jobs, privTxJob, markingTxJob)
	case txRequest.Params.Protocol == entities.GoQuorumChainType:
		privTxJob := uc.newJobEntityFromTxRequest(txRequest, uc.newEthTransactionFromParams(txRequest.Params, txData, entities.LegacyTxType),
			entities.TesseraPrivateTransaction, chainUUID)

		markingTx := &entities.ETHTransaction{
			From:         nil,
			PrivateFor:   txRequest.Params.PrivateFor,
			MandatoryFor: txRequest.Params.MandatoryFor,
			PrivacyFlag:  txRequest.Params.PrivacyFlag,
		}
		if txRequest.Params.From != nil {
			markingTx.From = txRequest.Params.From
		}
		markingTxJob := uc.newJobEntityFromTxRequest(txRequest, markingTx, entities.TesseraMarkingTransaction, chainUUID)
		jobs = append(jobs, privTxJob, markingTxJob)
	case txRequest.Params.Raw != nil:
		rawTx, err := uc.newTransactionFromRaw(txRequest.Params.Raw)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, uc.newJobEntityFromTxRequest(txRequest, rawTx, entities.EthereumRawTransaction, chainUUID))
	default:
		tx := uc.newEthTransactionFromParams(txRequest.Params, txData, txRequest.Params.TransactionType)
		jobs = append(jobs, uc.newJobEntityFromTxRequest(txRequest, tx, entities.EthereumTransaction, chainUUID))
	}

	return jobs, nil
}

func (uc *sendTxUsecase) newEthTransactionFromParams(params *entities.TxRequestParams, txData []byte, txType entities.TransactionType) *entities.ETHTransaction {
	tx := &entities.ETHTransaction{
		From:            nil,
		To:              nil,
		Nonce:           params.Nonce,
		Value:           params.Value,
		GasPrice:        params.GasPrice,
		Gas:             params.Gas,
		GasFeeCap:       params.GasFeeCap,
		GasTipCap:       params.GasTipCap,
		AccessList:      params.AccessList,
		TransactionType: txType,
		Raw:             params.Raw,
		Data:            txData,
		PrivateFrom:     params.PrivateFrom,
		PrivateFor:      params.PrivateFor,
		MandatoryFor:    params.MandatoryFor,
		PrivacyFlag:     params.PrivacyFlag,
		PrivacyGroupID:  params.PrivacyGroupID,
	}
	if params.From != nil {
		tx.From = params.From
	}
	if params.To != nil {
		tx.To = params.To
	}
	return tx
}

func (uc *sendTxUsecase) newJobEntityFromTxRequest(txRequest *entities.TxRequest, ethTx *entities.ETHTransaction, jobType entities.JobType, chainUUID string) *entities.Job {
	internalData := *txRequest.InternalData
	return &entities.Job{
		ScheduleUUID: txRequest.Schedule.UUID,
		ChainUUID:    chainUUID,
		Type:         jobType,
		Labels:       txRequest.Labels,
		InternalData: &internalData,
		Transaction:  ethTx,
		TenantID:     txRequest.Schedule.TenantID,
		OwnerID:      txRequest.Schedule.OwnerID,
	}
}

func (uc *sendTxUsecase) newTransactionFromRaw(raw hexutil.Bytes) (*entities.ETHTransaction, error) {
	tx := &types.Transaction{}

	err := tx.UnmarshalBinary(raw)
	if err != nil {
		errMessage := "failed to unmarshal raw transaction"
		uc.logger.WithError(err).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage)
	}

	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		errMessage := "failed to get sender from raw transaction"
		uc.logger.WithError(err).Error(errMessage)
		return nil, errors.InvalidParameterError(errMessage)
	}

	jobTx := &entities.ETHTransaction{
		From:     &from,
		Data:     tx.Data(),
		Gas:      utils.ToPtr(tx.Gas()).(*uint64),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Value:    (*hexutil.Big)(tx.Value()),
		Nonce:    utils.ToPtr(tx.Gas()).(*uint64),
		Hash:     utils.ToPtr(tx.Hash()).(*ethcommon.Hash),
		Raw:      raw,
	}

	// If not contract creation
	if tx.To() != nil {
		jobTx.To = tx.To()
	}

	return jobTx, nil
}
