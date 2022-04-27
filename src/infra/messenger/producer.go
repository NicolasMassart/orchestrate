package messenger

import (
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=producer.go -destination=mocks/producer.go -package=mocks

type Producer interface {
	SendJobMessage(topic string, job *entities.Job, partitionKey string, userInfo *multitenancy.UserInfo) error
	Checker() error
}
