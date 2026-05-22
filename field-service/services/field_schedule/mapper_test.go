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
