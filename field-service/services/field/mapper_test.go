package services

import (
	"field-service/domain/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestToFieldResponseMapsFullField(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	field := models.Field{
		UUID:         id,
		Code:         "F1",
		Name:         "Field 1",
		PricePerHour: 100000,
		Images:       pq.StringArray{"image-1"},
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	response := toFieldResponse(field)

	assert.Equal(t, id, response.UUID)
	assert.Equal(t, "F1", response.Code)
	assert.Equal(t, "Field 1", response.Name)
	assert.Equal(t, 100000, response.PricePerHour)
	assert.Equal(t, []string{"image-1"}, response.Images)
	assert.Equal(t, &now, response.CreatedAt)
	assert.Equal(t, &now, response.UpdatedAt)
}

func TestToFieldSummaryResponsePreservesExistingSummaryShape(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	field := models.Field{
		UUID:         id,
		Code:         "F1",
		Name:         "Field 1",
		PricePerHour: 100000,
		Images:       pq.StringArray{"image-1"},
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	response := toFieldSummaryResponse(field)

	assert.Equal(t, id, response.UUID)
	assert.Empty(t, response.Code)
	assert.Equal(t, "Field 1", response.Name)
	assert.Equal(t, 100000, response.PricePerHour)
	assert.Equal(t, []string{"image-1"}, response.Images)
	assert.Nil(t, response.CreatedAt)
	assert.Nil(t, response.UpdatedAt)
}

func TestToFieldResponsesMapsFields(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	fields := []models.Field{
		{UUID: firstID, Code: "F1", Name: "Field 1"},
		{UUID: secondID, Code: "F2", Name: "Field 2"},
	}

	responses := toFieldResponses(fields)

	assert.Len(t, responses, 2)
	assert.Equal(t, firstID, responses[0].UUID)
	assert.Equal(t, "F1", responses[0].Code)
	assert.Equal(t, secondID, responses[1].UUID)
	assert.Equal(t, "F2", responses[1].Code)
}

func TestToFieldSummaryResponsesMapsSummaries(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	fields := []models.Field{
		{UUID: firstID, Code: "F1", Name: "Field 1"},
		{UUID: secondID, Code: "F2", Name: "Field 2"},
	}

	responses := toFieldSummaryResponses(fields)

	assert.Len(t, responses, 2)
	assert.Equal(t, firstID, responses[0].UUID)
	assert.Empty(t, responses[0].Code)
	assert.Equal(t, secondID, responses[1].UUID)
	assert.Empty(t, responses[1].Code)
}
