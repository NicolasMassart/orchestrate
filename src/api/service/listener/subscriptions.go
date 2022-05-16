package listener

import (
	"bytes"
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/service/types"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/consensys/orchestrate/src/infra/api"
)

const EventLogsMessageType = "event-logs"

type SubscriptionHandler struct {
	notifySubscriptionUC usecases.NotifyContractEventsUseCase
}

func NewSubscriptionRouter(notifySubscriptionUC usecases.NotifyContractEventsUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{
		notifySubscriptionUC: notifySubscriptionUC,
	}
}

func (r *SubscriptionHandler) HandleEventLogs(ctx context.Context, msg *entities.Message) error {
	req := &types.EventLogsMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
	if err != nil {
		return errors.InvalidFormatError("invalid event logs request type")
	}

	err = r.notifySubscriptionUC.Execute(ctx, req.ChainUUID, req.Address, req.EventLogs, multitenancy.UserInfoValue(ctx))
	return err
}
