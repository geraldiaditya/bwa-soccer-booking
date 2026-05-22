package services

import (
	"field-service/domain/dto"
	"field-service/domain/models"
)

func toTimeResponse(timeModel models.Time) dto.TimeResponse {
	return dto.TimeResponse{
		UUID:      timeModel.UUID,
		StartTime: timeModel.StartTime,
		EndTime:   timeModel.EndTime,
		CreatedAt: timeModel.CreatedAt,
		UpdatedAt: timeModel.UpdatedAt,
	}
}

func toTimeResponses(times []models.Time) []dto.TimeResponse {
	responses := make([]dto.TimeResponse, 0, len(times))
	for _, timeModel := range times {
		responses = append(responses, toTimeResponse(timeModel))
	}
	return responses
}
