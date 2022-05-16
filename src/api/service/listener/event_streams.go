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

const SuspendEventStreamMessageType = "suspend-event-stream"

type EventStreamHandler struct {
	updateEventStreamUC usecases.UpdateEventStreamUseCase
}

func NewEventStreamHandler(updateEventStreamUC usecases.UpdateEventStreamUseCase) *EventStreamHandler {
	return &EventStreamHandler{
		updateEventStreamUC: updateEventStreamUC,
	}
}

func (r *EventStreamHandler) HandleEventStreamSuspend(ctx context.Context, msg *entities.Message) error {
	req := &types.SuspendEventStreamRequestMessage{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
	if err != nil {
		return errors.InvalidFormatError("invalid event stream suspend request type")
	}

	userInfo := multitenancy.UserInfoValue(ctx)

	_, err = r.updateEventStreamUC.Execute(ctx, &entities.EventStream{
		UUID:   req.UUID,
		Status: entities.EventStreamStatusSuspend,
	}, userInfo)

	return err
}
