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

const UpdateJobMessageType entities.RequestMessageType = "update-job"

type JobHandler struct {
	updateJobUC usecases.UpdateJobUseCase
}

func NewJobHandler(updateJobUC usecases.UpdateJobUseCase) *JobHandler {
	return &JobHandler{
		updateJobUC: updateJobUC,
	}
}

func (r *JobHandler) HandleJobUpdate(ctx context.Context, msg *entities.Message) error {
	req := &types.JobUpdateMessageRequest{}
	err := api.UnmarshalBody(bytes.NewReader(msg.Body), req)
	if err != nil {
		return errors.InvalidFormatError("invalid job update request type")
	}

	userInfo := multitenancy.UserInfoValue(ctx)
	_, err = r.updateJobUC.Execute(ctx, req.ToJobEntity(), req.Status, req.Message, userInfo)
	if err != nil && errors.IsInvalidStateError(err) {
		return nil
	}
	return err
}
