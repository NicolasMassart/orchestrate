package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/kafka/proto"
)

//go:generate mockgen -source=kafka.go -destination=mocks/kafka.go -package=mocks

type Client interface {
	Checker() error
}

type Producer interface {
	SendJobMessage(topic string, job *entities.Job, userInfo *multitenancy.UserInfo) error
	SendTxResponse(topic string, txResponse *proto.TxResponse) error
	Client() Client
}

type Consumer interface {
	Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	Close() error
	Client() Client
}
