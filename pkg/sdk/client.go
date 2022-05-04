package sdk

import (
	"context"

	utilstypes "github.com/consensys/quorum-key-manager/src/utils/api/types"

	qkmtypes "github.com/consensys/quorum-key-manager/src/stores/api/types"

	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	ethcommon "github.com/ethereum/go-ethereum/common"
	healthz "github.com/heptiolabs/healthcheck"
	dto "github.com/prometheus/client_model/go"
)

//go:generate mockgen -source=client.go -destination=mock/client.go -package=mock

type OrchestrateClient interface {
	TransactionClient
	ScheduleClient
	JobClient
	MetricClient
	AccountClient
	FaucetClient
	ChainClient
	ContractClient
	ChainProxyClient
	EventStreamClient
}

type ChainProxyClient interface {
	ChainProxyURL(chainUUID string) string
	ChainTesseraProxyURL(chainUUID string) string
}

type TransactionClient interface {
	SendContractTransaction(ctx context.Context, request *types.SendTransactionRequest) (*types.TransactionResponse, error)
	SendDeployTransaction(ctx context.Context, request *types.DeployContractRequest) (*types.TransactionResponse, error)
	SendRawTransaction(ctx context.Context, request *types.RawTransactionRequest) (*types.TransactionResponse, error)
	SendTransferTransaction(ctx context.Context, request *types.TransferRequest) (*types.TransactionResponse, error)
	GetTxRequest(ctx context.Context, txRequestUUID string) (*types.TransactionResponse, error)
	SendCallOffTransaction(ctx context.Context, txRequestUUID string) (*types.TransactionResponse, error)
	SendSpeedUpTransaction(ctx context.Context, txRequestUUID string, increment *float64) (*types.TransactionResponse, error)
}

type ScheduleClient interface {
	GetSchedule(ctx context.Context, scheduleUUID string) (*types.ScheduleResponse, error)
	GetSchedules(ctx context.Context) ([]*types.ScheduleResponse, error)
	CreateSchedule(ctx context.Context, request *types.CreateScheduleRequest) (*types.ScheduleResponse, error)
}

type JobClient interface {
	GetJob(ctx context.Context, jobUUID string) (*types.JobResponse, error)
	GetJobs(ctx context.Context) ([]*types.JobResponse, error)
	CreateJob(ctx context.Context, request *types.CreateJobRequest) (*types.JobResponse, error)
	UpdateJob(ctx context.Context, jobUUID string, request *types.UpdateJobRequest) (*types.JobResponse, error)
	StartJob(ctx context.Context, jobUUID string) error
	ResendJobTx(ctx context.Context, jobUUID string) error
	SearchJob(ctx context.Context, filters *entities.JobFilters) ([]*types.JobResponse, error)
}

type MetricClient interface {
	Checker() healthz.Check
	Prometheus(context.Context) (map[string]*dto.MetricFamily, error)
}

type AccountClient interface {
	CreateAccount(ctx context.Context, request *types.CreateAccountRequest) (*types.AccountResponse, error)
	SearchAccounts(ctx context.Context, filters *entities.AccountFilters) ([]*types.AccountResponse, error)
	GetAccount(ctx context.Context, address ethcommon.Address) (*types.AccountResponse, error)
	ImportAccount(ctx context.Context, request *types.ImportAccountRequest) (*types.AccountResponse, error)
	UpdateAccount(ctx context.Context, address ethcommon.Address, request *types.UpdateAccountRequest) (*types.AccountResponse, error)
	SignMessage(ctx context.Context, address ethcommon.Address, request *qkmtypes.SignMessageRequest) (string, error)
	SignTypedData(ctx context.Context, address ethcommon.Address, request *qkmtypes.SignTypedDataRequest) (string, error)
	VerifyMessageSignature(ctx context.Context, request *utilstypes.VerifyRequest) error
	VerifyTypedDataSignature(ctx context.Context, request *utilstypes.VerifyTypedDataRequest) error
}

type FaucetClient interface {
	RegisterFaucet(ctx context.Context, request *types.RegisterFaucetRequest) (*types.FaucetResponse, error)
	UpdateFaucet(ctx context.Context, uuid string, request *types.UpdateFaucetRequest) (*types.FaucetResponse, error)
	GetFaucet(ctx context.Context, uuid string) (*types.FaucetResponse, error)
	SearchFaucets(ctx context.Context, filters *entities.FaucetFilters) ([]*types.FaucetResponse, error)
	DeleteFaucet(ctx context.Context, uuid string) error
}

type ChainClient interface {
	RegisterChain(ctx context.Context, request *types.RegisterChainRequest) (*types.ChainResponse, error)
	UpdateChain(ctx context.Context, uuid string, request *types.UpdateChainRequest) (*types.ChainResponse, error)
	GetChain(ctx context.Context, uuid string) (*types.ChainResponse, error)
	SearchChains(ctx context.Context, filters *entities.ChainFilters) ([]*types.ChainResponse, error)
	DeleteChain(ctx context.Context, uuid string) error
}

type ContractClient interface {
	RegisterContract(ctx context.Context, req *types.RegisterContractRequest) (*types.ContractResponse, error)
	DeregisterContract(ctx context.Context, name, tag string) error
	GetContract(ctx context.Context, name, tag string) (*types.ContractResponse, error)
	SearchContract(ctx context.Context, req *types.SearchContractRequest) (*types.ContractResponse, error)
	GetContractsCatalog(ctx context.Context) ([]string, error)
	GetContractTags(ctx context.Context, name string) ([]string, error)
	SetContractAddressCodeHash(ctx context.Context, address, chainID string, req *types.SetContractCodeHashRequest) error
	GetContractEvents(ctx context.Context, address, chainUUID string, req *types.GetContractEventsRequest) (*types.GetContractEventsBySignHashResponse, error)
}

type EventStreamClient interface {
	CreateWebhookEventStream(ctx context.Context, request *types.CreateWebhookEventStreamRequest) (*types.EventStreamResponse, error)
	CreateKafkaEventStream(ctx context.Context, request *types.CreateKafkaEventStreamRequest) (*types.EventStreamResponse, error)
	UpdateWebhookEventStream(ctx context.Context, uuid string, request *types.UpdateWebhookEventStreamRequest) (*types.EventStreamResponse, error)
	UpdateKafkaEventStream(ctx context.Context, uuid string, request *types.UpdateKafkaEventStreamRequest) (*types.EventStreamResponse, error)
	GetEventStream(ctx context.Context, uuid string) (*types.EventStreamResponse, error)
	SearchEventStreams(ctx context.Context, filters *entities.EventStreamFilters) ([]*types.EventStreamResponse, error)
	DeleteEventStream(ctx context.Context, uuid string) error
}
