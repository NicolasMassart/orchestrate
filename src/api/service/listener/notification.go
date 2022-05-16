package listener

import (
	"bytes"
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/api"
)

const AckNotificationMessageType = "ack-notification"

type NotificationHandler struct {
	ackNotifUC usecases.AckNotificationUseCase
}

func NewNotificationHandler(ackNotifUC usecases.AckNotificationUseCase) *NotificationHandler {
	return &NotificationHandler{
		ackNotifUC: ackNotifUC,
	}
}

func (r *NotificationHandler) HandleNotificationAck(ctx context.Context, msg *entities.Message) error {
	req := &types.AckNotificationRequestMessage{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
	if err != nil {
		return errors.InvalidFormatError("invalid notification ack request type")
	}

	return r.ackNotifUC.Execute(ctx, req.UUID)
}
