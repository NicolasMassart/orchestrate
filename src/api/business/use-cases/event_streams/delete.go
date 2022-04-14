package streams

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
)

const deleteEventStreamComponent = "use-cases.delete-event_stream"

type deleteUseCase struct {
	db     store.EventStreamAgent
	logger *log.Logger
}

func NewDeleteUseCase(db store.EventStreamAgent) usecases.DeleteEventStreamUseCase {
	return &deleteUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(deleteEventStreamComponent),
	}
}

func (uc *deleteUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) error {
	ctx = log.WithFields(ctx, log.Field("event_stream_uuid", uuid))
	logger := uc.logger.WithContext(ctx)
	logger.Debug("deleting event stream")

	_, err := uc.db.FindOneByUUID(ctx, uuid, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return errors.FromError(err).ExtendComponent(deleteEventStreamComponent)
	}

	err = uc.db.Delete(ctx, uuid, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return errors.FromError(err).ExtendComponent(deleteEventStreamComponent)
	}

	logger.Info("event stream was deleted successfully")
	return nil
}
