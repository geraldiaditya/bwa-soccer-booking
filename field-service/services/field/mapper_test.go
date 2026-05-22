package services

import (
	"field-service/domain/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToFieldResponseMapsCompleteFieldForAPIResponse(t *testing.T) {
	id := uuid.New()
	createdAt := time.Date(2026, 5, 22, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	field := models.Field{
		UUID:         id,
		Code:         "F1",
		Name:         "Main Field",
		PricePerHour: 150000,
		Images:       []string{"main-field.jpg"},
		CreatedAt:    &createdAt,
		UpdatedAt:    &updatedAt,
	}

	response := toFieldResponse(field)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "F1", response.Code)
	assert.Equal(t, "Main Field", response.Name)
	assert.Equal(t, 150000, response.PricePerHour)
	assert.Equal(t, []string{"main-field.jpg"}, response.Images)
	assert.Equal(t, &createdAt, response.CreatedAt)
	assert.Equal(t, &updatedAt, response.UpdatedAt)
}

func TestToFieldSummaryResponseLeavesInternalCodeOutOfPublicList(t *testing.T) {
	field := models.Field{
		UUID:         uuid.New(),
		Code:         "PRIVATE-CODE",
		Name:         "Training Field",
		PricePerHour: 90000,
		Images:       []string{"training.jpg"},
	}

	response := toFieldSummaryResponse(field)

	assert.Equal(t, "Training Field", response.Name)
	assert.Equal(t, 90000, response.PricePerHour)
	assert.Equal(t, []string{"training.jpg"}, response.Images)
	assert.Empty(t, response.Code)
}

func TestToFieldResponsesMapsFields(t *testing.T) {
	fields := []models.Field{
		{UUID: uuid.New(), Code: "F1", Name: "Main Field", PricePerHour: 150000},
		{UUID: uuid.New(), Code: "F2", Name: "Training Field", PricePerHour: 90000},
	}

	responses := toFieldResponses(fields)

	assert.Len(t, responses, 2)
	assert.Equal(t, fields[0].UUID, responses[0].UUID)
	assert.Equal(t, "F1", responses[0].Code)
	assert.Equal(t, "Main Field", responses[0].Name)
	assert.Equal(t, fields[1].UUID, responses[1].UUID)
	assert.Equal(t, "F2", responses[1].Code)
	assert.Equal(t, "Training Field", responses[1].Name)
}

func TestToFieldSummaryResponsesMapsSummaries(t *testing.T) {
	fields := []models.Field{
		{UUID: uuid.New(), Code: "F1", Name: "Main Field", PricePerHour: 150000},
		{UUID: uuid.New(), Code: "F2", Name: "Training Field", PricePerHour: 90000},
	}

	responses := toFieldSummaryResponses(fields)

	assert.Len(t, responses, 2)
	assert.Equal(t, fields[0].UUID, responses[0].UUID)
	assert.Equal(t, "Main Field", responses[0].Name)
	assert.Empty(t, responses[0].Code)
	assert.Equal(t, fields[1].UUID, responses[1].UUID)
	assert.Equal(t, "Training Field", responses[1].Name)
	assert.Empty(t, responses[1].Code)
}
