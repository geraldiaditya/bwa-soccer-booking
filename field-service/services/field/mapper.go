package services

import (
	"field-service/domain/dto"
	"field-service/domain/models"
)

func toFieldResponse(field models.Field) dto.FieldResponse {
	return dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}
}

func toFieldSummaryResponse(field models.Field) dto.FieldResponse {
	return dto.FieldResponse{
		UUID:         field.UUID,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
	}
}

func toFieldResponses(fields []models.Field) []dto.FieldResponse {
	responses := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		responses = append(responses, toFieldResponse(field))
	}
	return responses
}

func toFieldSummaryResponses(fields []models.Field) []dto.FieldResponse {
	responses := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		responses = append(responses, toFieldSummaryResponse(field))
	}
	return responses
}
