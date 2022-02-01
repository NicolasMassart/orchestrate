// +build unit

package parsers

import (
	"github.com/consensys/orchestrate/src/entities/testdata"
	"testing"

	"github.com/stretchr/testify/assert"
	modelstestdata "github.com/consensys/orchestrate/src/api/store/models/testdata"

	"encoding/json"
)

func TestParsersSchedule_NewModelFromEntity(t *testing.T) {
	scheduleEntity := testdata.FakeSchedule()
	scheduleModel := NewScheduleModel(scheduleEntity)
	finalScheduleEntity := NewScheduleEntity(scheduleModel)

	expectedJSON, _ := json.Marshal(scheduleEntity)
	actualJOSN, _ := json.Marshal(finalScheduleEntity)
	assert.Equal(t, string(expectedJSON), string(actualJOSN))
}

func TestParsersSchedule_NewEntityFromModel(t *testing.T) {
	scheduleModel := modelstestdata.FakeSchedule("", "")
	scheduleEntity := NewScheduleEntity(scheduleModel)
	finalScheduleModel := NewScheduleModel(scheduleEntity)

	assert.Equal(t, finalScheduleModel.UUID, scheduleModel.UUID)
	assert.Equal(t, finalScheduleModel.TenantID, scheduleModel.TenantID)
	assert.Equal(t, finalScheduleModel.CreatedAt, scheduleModel.CreatedAt)
	assert.Equal(t, finalScheduleModel.Jobs[0].UUID, scheduleModel.Jobs[0].UUID)
	assert.Equal(t, finalScheduleModel.Jobs[0].Type, scheduleModel.Jobs[0].Type)
	assert.Equal(t, finalScheduleModel.Jobs[0].Labels, scheduleModel.Jobs[0].Labels)
}
