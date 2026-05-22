package services

import (
	"field-service/constants"
	"field-service/domain/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToFieldScheduleResponseMapsSchedule(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	schedule := models.FieldSchedule{
		UUID:      id,
		Date:      time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
		Status:    constants.Available,
		CreatedAt: &now,
		UpdatedAt: &now,
		Field: models.Field{
			Name:         "Field 1",
			PricePerHour: 100000,
		},
		Time: models.Time{
			StartTime: "08:00",
			EndTime:   "09:00",
		},
	}

	response := toFieldScheduleResponse(schedule)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "Field 1", response.FieldName)
	assert.Equal(t, 100000, response.PricePerHour)
	assert.Equal(t, "2026-05-22", response.Date)
	assert.Equal(t, constants.Available.GetStatusString(), response.Status)
	assert.Equal(t, "08:00 - 09:00", response.Time)
	assert.Equal(t, &now, response.CreatedAt)
	assert.Equal(t, &now, response.UpdatedAt)
}

func TestToFieldScheduleForBookingResponsePreservesBookingShape(t *testing.T) {
	id := uuid.New()
	schedule := models.FieldSchedule{
		UUID:   id,
		Date:   time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
		Status: constants.Available,
		Field: models.Field{
			PricePerHour: 100000,
		},
		Time: models.Time{
			StartTime: "08:00",
			EndTime:   "09:00",
		},
	}

	response := toFieldScheduleForBookingResponse(schedule)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "22 Mei", response.Date)
	assert.Equal(t, constants.Available.GetStatusString(), response.Status)
	assert.Equal(t, "08:00", response.Time)
	assert.Equal(t, "Rp. 100.000", response.PricePerHour)
}

func TestToFieldScheduleResponsesMapsSchedules(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	schedules := []models.FieldSchedule{
		{
			UUID:   firstID,
			Date:   time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
			Status: constants.Available,
			Field:  models.Field{Name: "Field 1", PricePerHour: 100000},
			Time:   models.Time{StartTime: "08:00", EndTime: "09:00"},
		},
		{
			UUID:   secondID,
			Date:   time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC),
			Status: constants.Booked,
			Field:  models.Field{Name: "Field 2", PricePerHour: 200000},
			Time:   models.Time{StartTime: "09:00", EndTime: "10:00"},
		},
	}

	responses := toFieldScheduleResponses(schedules)

	assert.Len(t, responses, 2)
	assert.Equal(t, firstID, responses[0].UUID)
	assert.Equal(t, "08:00 - 09:00", responses[0].Time)
	assert.Equal(t, secondID, responses[1].UUID)
	assert.Equal(t, "09:00 - 10:00", responses[1].Time)
}

func TestToFieldScheduleResponseWithTimeOverridesScheduleTime(t *testing.T) {
	schedule := models.FieldSchedule{
		UUID:   uuid.New(),
		Date:   time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
		Status: constants.Available,
		Field:  models.Field{Name: "Field 1", PricePerHour: 100000},
		Time:   models.Time{StartTime: "08:00", EndTime: "09:00"},
	}
	scheduleTime := models.Time{StartTime: "10:00", EndTime: "11:00"}

	response := toFieldScheduleResponseWithTime(schedule, scheduleTime)

	assert.Equal(t, "10:00 - 11:00", response.Time)
}
