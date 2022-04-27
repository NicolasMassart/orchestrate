package e2e

import (
	"context"
	"fmt"
	"net/http"
	"time"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	pkghttp "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/notifier/kafka/testutils"
	"github.com/consensys/orchestrate/tests/config"
	"github.com/consensys/orchestrate/tests/pkg/jwt"
	pkgutils "github.com/consensys/orchestrate/tests/pkg/utils"
	"github.com/spf13/viper"
)

type Environment struct {
	Logger               *log.Logger
	HTTPClient           *http.Client
	EthClient            *rpc.Client
	Client               orchestrateclient.OrchestrateClient
	ConsumerTracker      *testutils.NotifierConsumerTracker
	KafkaTopic           string
	TestData             *config.TestDataCfg
	Artifacts            map[string]*config.Artifact
	UserInfo             *multitenancy.UserInfo
	WaitForTxResponseTTL time.Duration
	ctx                  context.Context
	ctxCancel            context.CancelFunc
	streamUUIDs          []string
	chainUUIDs           []string
}

func NewEnvironment(ctx context.Context, ctxCancel context.CancelFunc) (*Environment, error) {
	config.EnvInit()

	cfg, err := config.NewE2eConfig(viper.GetViper())
	if err != nil {
		return nil, err
	}

	consumerID := utils.RandString(10)
	cfg.KafkaCfg.ClientID = "client-" + consumerID
	cfg.KafkaCfg.GroupName = "group-" + consumerID

	logger := log.NewLogger()
	httpCfg := pkghttp.NewDefaultConfig()

	waitForTxResponse, err := time.ParseDuration(cfg.TestData.Timeout)
	if err != nil {
		return nil, err
	}

	userInfo := multitenancy.NewUserInfo(multitenancy.DefaultTenant, "")
	if cfg.TestData.OIDC != nil {
		jwtToken, err2 := jwt.GenerateJWT(cfg.TestData.OIDC.TokenURL, cfg.TestData.OIDC.ClientID, cfg.TestData.OIDC.ClientSecret, cfg.TestData.OIDC.Audience)
		if err2 != nil {
			return nil, fmt.Errorf("failed to generate JWT. %s", err2.Error())
		}

		httpCfg.Authorization = "Bearer " + jwtToken
		logger.Debug("Running using provided OIDC credentials")
	}

	if cfg.TestData.UserInfo != nil {
		userInfo = multitenancy.NewUserInfo(cfg.TestData.UserInfo.TenantID, cfg.TestData.UserInfo.Username)
	}

	httpClient := pkghttp.NewClient(httpCfg)
	orchClient := orchestrateclient.NewHTTPClient(httpClient, cfg.OrchestrateCfg)
	ethClient := rpc.NewClient(httpClient)

	kafkaTopic := "test-topic-" + utils.RandString(5)
	consumerTracker, err := testutils.NewNotifierConsumerTracker(cfg.KafkaCfg, []string{kafkaTopic})
	if err != nil {
		return nil, err
	}

	return &Environment{
		ctx:                  ctx,
		ctxCancel:            ctxCancel,
		Logger:               logger,
		Client:               orchClient,
		HTTPClient:           httpClient,
		EthClient:            ethClient,
		ConsumerTracker:      consumerTracker,
		KafkaTopic:           kafkaTopic,
		TestData:             cfg.TestData,
		Artifacts:            cfg.Artifacts,
		UserInfo:             userInfo,
		WaitForTxResponseTTL: waitForTxResponse,
		streamUUIDs:          []string{},
		chainUUIDs:           []string{},
	}, nil
}

func (env *Environment) Start() error {
	go func() {
		err := env.ConsumerTracker.Consume(env.ctx)
		if err != nil && env.ctx.Err() == nil {
			env.Logger.WithError(err).Error("failed to start kafka consumer")
			env.ctxCancel()
		}
	}()

	// TODO: Wait for consumer to be ready in a synchronous way
	env.Logger.Info("waiting consumer to be ready")
	time.Sleep(time.Second * 5)
	return nil
}

func (env *Environment) Stop() error {
	errs := []error{}
	for _, streamUUID := range env.streamUUIDs {
		err := env.Client.DeleteEventStream(env.ctx, streamUUID)
		if err != nil {
			env.Logger.WithError(err).Error("failed to delete event stream")
			errs = append(errs, err)
		}
	}

	err := env.ConsumerTracker.Close()
	if err != nil {
		env.Logger.WithError(err).Error("failed to stop consumer")
		errs = append(errs, err)
	}

	for _, chainUUID := range env.chainUUIDs {
		err := env.Client.DeleteChain(env.ctx, chainUUID)
		if err != nil {
			env.Logger.WithError(err).Error("failed to delete chain")
			errs = append(errs, err)
		}
	}

	env.ctxCancel()
	env.Logger.Info("setup test teardown has completed")
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (env *Environment) createChainWithStream(chainName string, urls []string, privateTxManagerURL string) (*entities.Chain, string, error) {
	chainRes, err := pkgutils.RegisterChainAndWaitForProxy(env.ctx, env.Client, env.EthClient, &types.RegisterChainRequest{
		Name:                chainName,
		URLs:                urls,
		PrivateTxManagerURL: privateTxManagerURL,
	})
	if err != nil {
		return nil, "", err
	}
	env.chainUUIDs = append(env.chainUUIDs, chainRes.UUID)
	env.Logger.WithField("chain_name", chainName).WithField("chain_uuid", chainRes.UUID).
		Info("chain created successfully")

	streamResp, err := env.Client.CreateKafkaEventStream(env.ctx, &types.CreateKafkaEventStreamRequest{
		Name:  "tx-stream-" + chainName,
		Topic: env.KafkaTopic,
		Chain: chainName,
	})
	if err != nil {
		return nil, "", err
	}
	env.streamUUIDs = append(env.streamUUIDs, streamResp.UUID)
	env.Logger.WithField("chain_name", chainName).WithField("stream_uuid", streamResp.UUID).
		Info("event stream created successfully")

	return formatters.ChainResponseToEntity(chainRes), streamResp.UUID, nil
}
