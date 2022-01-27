// +build unit

package api

import (
	ethclientmock "github.com/consensys/orchestrate/src/infra/ethclient/mock"
	mocks2 "github.com/consensys/quorum-key-manager/pkg/client/mock"

	"testing"

	"github.com/consensys/orchestrate/src/infra/broker/sarama"

	"github.com/Shopify/sarama/mocks"

	mockauth "github.com/consensys/orchestrate/pkg/toolkit/app/auth/mock"
	"github.com/consensys/orchestrate/src/infra/database/postgres"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := NewConfig(viper.New())
	cfg.Store.Type = "postgres"

	kCfg := sarama.NewKafkaTopicConfig(viper.New())
	_, err := NewAPI(
		cfg,
		postgres.GetManager(),
		mockauth.NewMockChecker(ctrl), mockauth.NewMockChecker(ctrl),
		mocks2.NewMockKeyManagerClient(ctrl),
		"defaultStoreID",
		ethclientmock.NewMockClient(ctrl),
		mocks.NewSyncProducer(t, nil),
		kCfg,
	)
	assert.NoError(t, err, "Creating App should not error")
}
