package e2e

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pkghttp "github.com/consensys/orchestrate/pkg/toolkit/app/http"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	"github.com/consensys/orchestrate/src/infra/kafka/testutils"
	"github.com/consensys/orchestrate/tests/pkg/jwt"

	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/tests/config"
	"github.com/spf13/viper"
)

type Environment struct {
	Logger               *log.Logger
	HTTPClient           *http.Client
	EthClient            *rpc.Client
	Client               orchestrateclient.OrchestrateClient
	KafkaConsumer        *testutils.ExternalConsumerTracker
	TestData             *config.TestDataCfg
	Artifacts            map[string]*config.Artifact
	UserInfo             *multitenancy.UserInfo
	WaitForTxResponseTTL time.Duration
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
			return nil, fmt.Errorf("faield to generate JWT. %s", err2.Error())
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

	consumerTracker, err := testutils.NewExternalConsumerTracker(cfg.KafkaCfg)
	if err != nil {
		return nil, err
	}

	go func() {
		err = consumerTracker.Consume(ctx, []string{"topic-tx-decoded"})
		if err != nil && ctx.Err() == nil {
			logger.WithError(err).Error("failed to start kafka consumer")
			ctxCancel()
		}
	}()

	// TODO: Wait for consumer to be ready in a synchronous way
	time.Sleep(time.Second * 5)

	return &Environment{
		Logger:               logger,
		Client:               orchClient,
		HTTPClient:           httpClient,
		EthClient:            ethClient,
		KafkaConsumer:        consumerTracker,
		TestData:             cfg.TestData,
		Artifacts:            cfg.Artifacts,
		UserInfo:             userInfo,
		WaitForTxResponseTTL: waitForTxResponse,
	}, nil
}
