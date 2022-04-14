package streams

import (
	"context"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/toolkit/app/log"
	"github.com/consensys/orchestrate/pkg/toolkit/app/multitenancy"
	usecases "github.com/consensys/orchestrate/src/api/business/use-cases"
	"github.com/consensys/orchestrate/src/api/store"
	"github.com/consensys/orchestrate/src/entities"
)

const getEventStreamComponent = "use-cases.get-event_stream"

type getUseCase struct {
	db     store.EventStreamAgent
	logger *log.Logger
}

func NewGetUseCase(db store.EventStreamAgent) usecases.GetEventStreamUseCase {
	return &getUseCase{
		db:     db,
		logger: log.NewLogger().SetComponent(getEventStreamComponent),
	}
}

func (uc *getUseCase) Execute(ctx context.Context, uuid string, userInfo *multitenancy.UserInfo) (*entities.EventStream, error) {
	ctx = log.WithFields(ctx, log.Field("event_stream_uuid", uuid))
	logger := uc.logger.WithContext(ctx)

	e, err := uc.db.FindOneByUUID(ctx, uuid, userInfo.AllowedTenants, userInfo.Username)
	if err != nil {
		return nil, errors.FromError(err).ExtendComponent(getEventStreamComponent)
	}

	logger.WithField("name", e.Name).Debug("event stream found successfully")
	return e, nil
}
