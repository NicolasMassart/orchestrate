package usecases

import (
	"context"

	"github.com/consensys/orchestrate/src/entities"
)

//go:generate mockgen -source=notifications.go -destination=mocks/notifications.go -package=mocks

type NotificationUseCases interface {
	NotifyFailedJob() NotifyFailedJob
	NotifyMinedJob() NotifyMinedJob
}

type NotifyFailedJob interface {
	Execute(ctx context.Context, job *entities.Job, errMsg string) error
}

type NotifyMinedJob interface {
	Execute(ctx context.Context, job *entities.Job) error
}
