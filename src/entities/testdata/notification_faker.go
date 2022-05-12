package testdata

import (
	"github.com/consensys/orchestrate/src/entities"
	"github.com/gofrs/uuid"
)

func FakeNotification() *entities.Notification {
	fakeJob := FakeJob()
	return &entities.Notification{
		UUID:       uuid.Must(uuid.NewV4()).String(),
		SourceUUID: fakeJob.UUID,
		SourceType: entities.NotificationSourceTypeJob,
		Status:     entities.NotificationStatusPending,
		Type:       entities.NotificationTypeTxFailed,
		APIVersion: "v1",
		Job:        fakeJob,
		Error:      "error message",
	}
}
