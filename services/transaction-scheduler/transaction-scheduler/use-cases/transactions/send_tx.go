package transactions

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/database"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/multitenancy"
	types "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/chainregistry"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/entities"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/utils"
	chainregistry "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/client"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/store"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/store/models"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/parsers"
	usecases "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/use-cases"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/transaction-scheduler/transaction-scheduler/validators"
)

const sendTxComponent = "use-cases.send-tx"

// sendTxUsecase is a use case to create a new transaction
type sendTxUsecase struct {
	validator           validators.TransactionValidator
	db                  store.DB
	chainRegistryClient chainregistry.ChainRegistryClient
	startJobUC          usecases.StartJobUseCase
	createJobUC         usecases.CreateJobUseCase
	getTxUC             usecases.GetTxUseCase
}

// NewSendTxUseCase creates a new SendTxUseCase
func NewSendTxUseCase(
	validator validators.TransactionValidator,
	db store.DB,
	chainRegistryClient chainregistry.ChainRegistryClient,
	startJobUseCase usecases.StartJobUseCase,
	createJobUC usecases.CreateJobUseCase,
	getTxUC usecases.GetTxUseCase,
) usecases.SendTxUseCase {
	return &sendTxUsecase{
		validator:           validator,
		db:                  db,
		chainRegistryClient: chainRegistryClient,
		startJobUC:          startJobUseCase,
		createJobUC:         createJobUC,
		getTxUC:             getTxUC,
	}
}

// Execute validates, creates and starts a new transaction
func (uc *sendTxUsecase) Execute(ctx context.Context, txRequest *entities.TxRequest, txData, tenantID string) (*entities.TxRequest, error) {
	logger := log.WithContext(ctx).WithField("idempotency_key", txRequest.IdempotencyKey)
	logger.Debug("creating new transaction")

	// Step 1: Get chainUUID from chain registry
	chainUUID, err := uc.getChainUUID(ctx, txRequest.ChainName)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 2: Generate request hash
	requestHash, err := generateRequestHash(chainUUID, txRequest.Params)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 3: Insert Schedule + Job + Transaction + TxRequest atomically OR get tx request if it exists
	txRequest, err = uc.selectOrInsertTxRequest(ctx, txRequest, txData, requestHash, chainUUID, tenantID)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 4: Start first job of the schedule if status is CREATED
	// Otherwise there was another request with same idempotency key and same reqHash
	job := txRequest.Schedule.Jobs[0]
	if job.Status == utils.StatusCreated {
		var fctJob *entities.Job
		fctJob, err = uc.startFaucetJob(ctx, txRequest.Params.From, chainUUID, job.ScheduleUUID, tenantID)
		if err != nil {
			logger.WithError(err).Error("could not start faucet job")
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
		if fctJob != nil {
			txRequest.Schedule.Jobs = append(txRequest.Schedule.Jobs, fctJob)
		}

		err = uc.startJobUC.Execute(ctx, job.UUID, []string{tenantID})
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
	} else { // Load latest Schedule status from DB
		txRequest, err = uc.getTxUC.Execute(ctx, txRequest.Schedule.UUID, []string{tenantID})
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
	}

	logger.WithField("schedule.uuid", txRequest.Schedule.UUID).Info("send transaction request created successfully")
	return txRequest, nil
}

func (uc *sendTxUsecase) getChainUUID(ctx context.Context, chainName string) (string, error) {
	chain, err := uc.chainRegistryClient.GetChainByName(ctx, chainName)
	if err != nil {
		errMessage := fmt.Sprintf("cannot load '%s' chain", chainName)
		log.WithContext(ctx).WithError(err).Error(errMessage)
		if errors.IsNotFoundError(err) {
			return "", errors.InvalidParameterError(errMessage)
		}
		return "", errors.FromError(err)
	}

	return chain.UUID, nil
}

func (uc *sendTxUsecase) selectOrInsertTxRequest(
	ctx context.Context,
	txRequest *entities.TxRequest,
	txData, requestHash, chainUUID, tenantID string,
) (*entities.TxRequest, error) {
	txRequestModel, err := uc.db.TransactionRequest().FindOneByIdempotencyKey(ctx, txRequest.IdempotencyKey, tenantID)
	switch {
	case errors.IsNotFoundError(err):
		return uc.insertNewTxRequest(ctx, txRequest, txData, requestHash, chainUUID, tenantID)
	case err != nil:
		return nil, err
	case txRequestModel != nil && txRequestModel.RequestHash != requestHash:
		errMessage := "a transaction request with the same idempotency key and different params already exists"
		log.WithError(err).WithField("idempotency_key", txRequestModel.IdempotencyKey).Error(errMessage)
		return nil, errors.AlreadyExistsError(errMessage)
	default:
		return uc.getTxUC.Execute(ctx, txRequestModel.Schedule.UUID, []string{tenantID})
	}
}

func (uc *sendTxUsecase) insertNewTxRequest(
	ctx context.Context,
	txRequest *entities.TxRequest,
	txData, requestHash, chainUUID, tenantID string,
) (*entities.TxRequest, error) {
	err := database.ExecuteInDBTx(uc.db, func(dbtx database.Tx) error {
		schedule := &models.Schedule{TenantID: tenantID}
		if err := dbtx.(store.Tx).Schedule().Insert(ctx, schedule); err != nil {
			return err
		}

		txRequestModel := parsers.NewTxRequestModelFromEntities(txRequest, requestHash, schedule.ID)
		if err := dbtx.(store.Tx).TransactionRequest().Insert(ctx, txRequestModel); err != nil {
			return err
		}

		txRequest.Schedule = parsers.NewScheduleEntityFromModels(schedule)
		return nil
	})
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	sendTxJobs := parsers.NewJobEntitiesFromTxRequest(txRequest, chainUUID, txData)
	txRequest.Schedule.Jobs = make([]*entities.Job, len(sendTxJobs))
	err = database.ExecuteInDBTx(uc.db, func(dbtx database.Tx) error {
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
			job, err = uc.createJobUC.WithDBTransaction(dbtx.(store.Tx)).
				Execute(ctx, txJob, []string{tenantID})
			if err != nil {
				return err
			}

			txRequest.Schedule.Jobs[idx] = job
		}
		return nil
	})

	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	return txRequest, err
}

// Execute validates, creates and starts a new transaction for pre funding users account
func (uc *sendTxUsecase) startFaucetJob(ctx context.Context, account, chainUUID, scheduleUUID, tenantID string) (*entities.Job, error) {
	if account == "" || account == ethcommon.HexToAddress("0x").String() {
		return nil, nil
	}

	logger := log.WithContext(ctx).WithField("chain_uuid", chainUUID)

	fct, err := uc.chainRegistryClient.GetFaucetCandidate(multitenancy.WithTenantID(ctx, tenantID), ethcommon.HexToAddress(account), chainUUID)
	if err != nil {
		return nil, err
	}
	if fct == nil {
		logger.Debug("could not find a candidate faucets")
		return nil, nil
	}

	logger.WithFields(log.Fields{
		"faucet.amount": fct.Amount,
	}).Infof("faucet: credit approved")

	txJob := generateFaucetJob(fct, scheduleUUID, chainUUID, account)

	fctJob, err := uc.createJobUC.Execute(ctx, txJob, []string{tenantID})
	if err != nil {
		return nil, err
	}

	err = uc.startJobUC.Execute(ctx, fctJob.UUID, []string{tenantID})
	if err != nil {
		return fctJob, err
	}

	return fctJob, nil
}

func generateFaucetJob(fct *types.Faucet, scheduleUUID, chainUUID, account string) *entities.Job {
	return &entities.Job{
		ScheduleUUID: scheduleUUID,
		ChainUUID:    chainUUID,
		Type:         utils.EthereumTransaction,
		Labels:       types.FaucetToJobLabels(fct),
		InternalData: &entities.InternalData{},
		Transaction: &entities.ETHTransaction{
			From:  fct.Creditor.Hex(),
			To:    account,
			Value: fct.Amount.String(),
		},
	}
}

func generateRequestHash(chainUUID string, params interface{}) (string, error) {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	hash := md5.Sum([]byte(string(jsonParams) + chainUUID))
	return hex.EncodeToString(hash[:]), nil
}
