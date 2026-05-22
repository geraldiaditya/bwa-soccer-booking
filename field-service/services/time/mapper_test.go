package services

import (
	"field-service/domain/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToTimeResponseMapsTime(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	timeModel := models.Time{
		UUID:      id,
		StartTime: "08:00",
		EndTime:   "09:00",
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	response := toTimeResponse(timeModel)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "08:00", response.StartTime)
	assert.Equal(t, "09:00", response.EndTime)
	assert.Equal(t, &now, response.CreatedAt)
	assert.Equal(t, &now, response.UpdatedAt)
}

func TestToTimeResponsesMapsTimes(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	times := []models.Time{
		{UUID: firstID, StartTime: "08:00", EndTime: "09:00"},
		{UUID: secondID, StartTime: "09:00", EndTime: "10:00"},
	}

	responses := toTimeResponses(times)

	assert.Len(t, responses, 2)
	assert.Equal(t, firstID, responses[0].UUID)
	assert.Equal(t, "08:00", responses[0].StartTime)
	assert.Equal(t, secondID, responses[1].UUID)
	assert.Equal(t, "10:00", responses[1].EndTime)
}
