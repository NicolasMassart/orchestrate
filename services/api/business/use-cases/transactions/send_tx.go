package transactions

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/chain-registry/store/models"

	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/database"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/multitenancy"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/types/entities"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/pkg/utils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/business/parsers"
	usecases "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/business/use-cases"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/business/validators"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/api/store"
	chainregistry "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/v2/services/chain-registry/client"
)

const sendTxComponent = "use-cases.send-tx"

// sendTxUsecase is a use case to create a new transaction
type sendTxUsecase struct {
	validator           validators.TransactionValidator
	db                  store.DB
	chainRegistryClient chainregistry.ChainRegistryClient
	startJobUC          usecases.StartJobUseCase
	createJobUC         usecases.CreateJobUseCase
	createScheduleUC    usecases.CreateScheduleUseCase
	getTxUC             usecases.GetTxUseCase
	getFaucetCandidate  usecases.GetFaucetCandidateUseCase
}

// NewSendTxUseCase creates a new SendTxUseCase
func NewSendTxUseCase(
	validator validators.TransactionValidator,
	db store.DB,
	chainRegistryClient chainregistry.ChainRegistryClient,
	startJobUseCase usecases.StartJobUseCase,
	createJobUC usecases.CreateJobUseCase,
	createScheduleUC usecases.CreateScheduleUseCase,
	getTxUC usecases.GetTxUseCase,
	getFaucetCandidate usecases.GetFaucetCandidateUseCase,
) usecases.SendTxUseCase {
	return &sendTxUsecase{
		validator:           validator,
		db:                  db,
		chainRegistryClient: chainRegistryClient,
		startJobUC:          startJobUseCase,
		createJobUC:         createJobUC,
		createScheduleUC:    createScheduleUC,
		getTxUC:             getTxUC,
		getFaucetCandidate:  getFaucetCandidate,
	}
}

// Execute validates, creates and starts a new transaction
func (uc *sendTxUsecase) Execute(ctx context.Context, txRequest *entities.TxRequest, txData, tenantID string) (*entities.TxRequest, error) {
	logger := log.WithContext(ctx).WithField("idempotency_key", txRequest.IdempotencyKey)
	logger.Debug("creating new transaction")

	// Step 1: Get chain from chain registry
	chain, err := uc.getChain(ctx, txRequest.ChainName)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 2: Generate request hash
	requestHash, err := generateRequestHash(chain.UUID, txRequest.Params)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 3: Insert Schedule + Job + Transaction + TxRequest atomically OR get tx request if it exists
	txRequest, err = uc.selectOrInsertTxRequest(ctx, txRequest, txData, requestHash, chain.UUID, tenantID)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	// Step 4: Start first job of the schedule if status is CREATED
	// Otherwise there was another request with same idempotency key and same reqHash
	job := txRequest.Schedule.Jobs[0]
	if job.Status == utils.StatusCreated {
		err = uc.startFaucetJob(ctx, txRequest.Params.From, job.ScheduleUUID, tenantID, chain)
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}

		err = uc.startJobUC.Execute(ctx, job.UUID, []string{tenantID})
		if err != nil {
			return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
		}
	}

	// Step 5: Load latest Schedule status from DB
	txRequest, err = uc.getTxUC.Execute(ctx, txRequest.Schedule.UUID, []string{tenantID})
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(sendTxComponent)
	}

	logger.WithField("schedule.uuid", txRequest.Schedule.UUID).Info("send transaction request created successfully")
	return txRequest, nil
}

func (uc *sendTxUsecase) getChain(ctx context.Context, chainName string) (*models.Chain, error) {
	chain, err := uc.chainRegistryClient.GetChainByName(ctx, chainName)
	if err != nil {
		errMessage := fmt.Sprintf("cannot load '%s' chain", chainName)
		log.WithContext(ctx).WithError(err).Error(errMessage)
		if errors.IsNotFoundError(err) {
			return nil, errors.InvalidParameterError(errMessage)
		}

		return nil, err
	}

	return chain, nil
}

func (uc *sendTxUsecase) selectOrInsertTxRequest(
	ctx context.Context,
	txRequest *entities.TxRequest,
	txData, requestHash, chainUUID, tenantID string,
) (*entities.TxRequest, error) {
	if txRequest.IdempotencyKey == "" {
		return uc.insertNewTxRequest(ctx, txRequest, txData, requestHash, chainUUID, tenantID)
	}

	txRequestModel, err := uc.db.TransactionRequest().FindOneByIdempotencyKey(ctx, txRequest.IdempotencyKey, tenantID)
	switch {
	case errors.IsNotFoundError(err):
		return uc.insertNewTxRequest(ctx, txRequest, txData, requestHash, chainUUID, tenantID)
	case err != nil:
		return nil, err
	case txRequestModel != nil && txRequestModel.RequestHash != requestHash:
		errMessage := "a transaction request with the same idempotency key and different params already exists"
		log.WithField("idempotency_key", txRequestModel.IdempotencyKey).Error(errMessage)
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
		schedule, der := uc.createScheduleUC.WithDBTransaction(dbtx.(store.Tx)).Execute(ctx, &entities.Schedule{TenantID: tenantID})
		if der != nil {
			return der
		}
		txRequest.Schedule = schedule

		scheduleModel, der := dbtx.(store.Tx).Schedule().FindOneByUUID(ctx, txRequest.Schedule.UUID, []string{tenantID})
		if der != nil {
			return der
		}

		txRequestModel := parsers.NewTxRequestModelFromEntities(txRequest, requestHash, scheduleModel.ID)
		der = dbtx.(store.Tx).TransactionRequest().Insert(ctx, txRequestModel)
		if der != nil {
			return der
		}

		sendTxJobs := parsers.NewJobEntitiesFromTxRequest(txRequest, chainUUID, txData)
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

			job, der := uc.createJobUC.WithDBTransaction(dbtx.(store.Tx)).Execute(ctx, txJob, []string{tenantID})
			if der != nil {
				return der
			}

			txRequest.Schedule.Jobs[idx] = job
		}

		return nil
	})

	return txRequest, err
}

// Execute validates, creates and starts a new transaction for pre funding users account
func (uc *sendTxUsecase) startFaucetJob(ctx context.Context, account, scheduleUUID, tenantID string, chain *models.Chain) error {
	if account == "" {
		return nil
	}

	logger := log.WithContext(ctx).WithField("chain_uuid", chain.UUID)

	faucet, err := uc.getFaucetCandidate.Execute(ctx, account, chain, []string{tenantID, multitenancy.DefaultTenant})
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil
		}

		return errors.FromError(err).ExtendComponent(sendTxComponent)
	}
	logger.WithField("faucet_amount", faucet.Amount).Info("faucet: credit approved")

	txJob := &entities.Job{
		ScheduleUUID: scheduleUUID,
		ChainUUID:    chain.UUID,
		Type:         utils.EthereumTransaction,
		Labels: map[string]string{
			"faucetUUID": faucet.UUID,
		},
		InternalData: &entities.InternalData{},
		Transaction: &entities.ETHTransaction{
			From:  faucet.CreditorAccount,
			To:    account,
			Value: faucet.Amount,
		},
	}
	fctJob, err := uc.createJobUC.Execute(ctx, txJob, []string{tenantID})
	if err != nil {
		return err
	}

	return uc.startJobUC.Execute(ctx, fctJob.UUID, []string{tenantID})
}

func generateRequestHash(chainUUID string, params interface{}) (string, error) {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	hash := md5.Sum([]byte(string(jsonParams) + chainUUID))
	return hex.EncodeToString(hash[:]), nil
}