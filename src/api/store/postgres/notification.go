package postgres

import (
	"context"
	"time"

	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/api/store/models"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/postgres"
	"github.com/gofrs/uuid"
)

type PGNotification struct {
	client postgres.Client
	logger *log.Logger
}

var _ store.NotificationAgent = &PGNotification{}

func NewPGNotification(client postgres.Client) *PGNotification {
	return &PGNotification{
		client: client,
		logger: log.NewLogger().SetComponent("data-agents.notification"),
	}
}

func (agent *PGNotification) Insert(ctx context.Context, notif *entities.Notification) (*entities.Notification, error) {
	model := models.NewNotification(notif)

	model.UUID = uuid.Must(uuid.NewV4()).String()
	model.CreatedAt = time.Now().UTC()
	model.UpdatedAt = model.CreatedAt

	err := agent.client.ModelContext(ctx, model).Insert()
	if err != nil {
		errMsg := "failed to insert notification"
		agent.logger.WithContext(ctx).WithError(err).Error(errMsg)
		return nil, errors.FromError(err).SetMessage(errMsg)
	}

	return model.ToEntity(), nil
}

func (agent *PGNotification) Update(ctx context.Context, notif *entities.Notification) (*entities.Notification, error) {
	model := models.NewNotification(notif)
	model.UpdatedAt = time.Now().UTC()

	err := agent.client.ModelContext(ctx, model).Where("uuid = ?", notif.UUID).UpdateNotZero()
	if err != nil {
		errMessage := "failed to update notification"
		agent.logger.WithContext(ctx).WithError(err).Error(errMessage)
		return nil, errors.FromError(err).SetMessage(errMessage)
	}

	return model.ToEntity(), nil
}
