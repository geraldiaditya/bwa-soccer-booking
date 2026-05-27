package services

import (
	"field-service/constants"
	"field-service/domain/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToFieldScheduleResponseMapsDisplayFieldsForAdminAPI(t *testing.T) {
	id := uuid.New()
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	schedule := models.FieldSchedule{
		UUID:   id,
		Date:   date,
		Status: constants.Available,
		Field: models.Field{
			Name:         "Main Field",
			PricePerHour: 150000,
		},
		Time: models.Time{StartTime: "08:00", EndTime: "09:00"},
	}

	response := toFieldScheduleResponse(schedule)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "Main Field", response.FieldName)
	assert.Equal(t, 150000, response.PricePerHour)
	assert.Equal(t, "2026-06-01", response.Date)
	assert.Equal(t, constants.Available.GetStatusString(), response.Status)
	assert.Equal(t, "08:00 - 09:00", response.Time)
}

func TestToFieldScheduleForBookingResponseFormatsCustomerFacingValues(t *testing.T) {
	id := uuid.New()
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	schedule := models.FieldSchedule{
		UUID:   id,
		Date:   date,
		Status: constants.Booked,
		Field:  models.Field{PricePerHour: 150000},
		Time:   models.Time{StartTime: "08:00", EndTime: "09:00"},
	}

	response := toFieldScheduleForBookingResponse(schedule)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "01 Jun", response.Date)
	assert.Equal(t, "08:00", response.Time)
	assert.Equal(t, constants.Booked.GetStatusString(), response.Status)
	assert.Equal(t, "Rp. 150.000", response.PricePerHour)
}

func TestToFieldScheduleForBookingResponseFormatsMayDateWithoutExtraSpace(t *testing.T) {
	date := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	schedule := models.FieldSchedule{
		UUID:   uuid.New(),
		Date:   date,
		Status: constants.Available,
		Field:  models.Field{PricePerHour: 100000},
		Time:   models.Time{StartTime: "10:00", EndTime: "11:00"},
	}

	response := toFieldScheduleForBookingResponse(schedule)

	assert.Equal(t, "22 Mei", response.Date)
}

func TestToFieldScheduleResponsesMapsSchedules(t *testing.T) {
	schedules := []models.FieldSchedule{
		{
			UUID:   uuid.New(),
			Date:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			Status: constants.Available,
			Field:  models.Field{Name: "Main Field", PricePerHour: 150000},
			Time:   models.Time{StartTime: "08:00", EndTime: "09:00"},
		},
		{
			UUID:   uuid.New(),
			Date:   time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
			Status: constants.Booked,
			Field:  models.Field{Name: "Training Field", PricePerHour: 90000},
			Time:   models.Time{StartTime: "09:00", EndTime: "10:00"},
		},
	}

	responses := toFieldScheduleResponses(schedules)

	assert.Len(t, responses, 2)
	assert.Equal(t, schedules[0].UUID, responses[0].UUID)
	assert.Equal(t, "Main Field", responses[0].FieldName)
	assert.Equal(t, "08:00 - 09:00", responses[0].Time)
	assert.Equal(t, schedules[1].UUID, responses[1].UUID)
	assert.Equal(t, "Training Field", responses[1].FieldName)
	assert.Equal(t, "09:00 - 10:00", responses[1].Time)
}

func TestToFieldScheduleResponseWithTimeOverridesScheduleTime(t *testing.T) {
	schedule := models.FieldSchedule{
		UUID:   uuid.New(),
		Date:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Status: constants.Available,
		Field:  models.Field{Name: "Main Field", PricePerHour: 150000},
		Time:   models.Time{StartTime: "08:00", EndTime: "09:00"},
	}
	updatedTime := models.Time{StartTime: "11:00", EndTime: "12:00"}

	response := toFieldScheduleResponseWithTime(schedule, updatedTime)

	assert.Equal(t, "11:00 - 12:00", response.Time)
}
