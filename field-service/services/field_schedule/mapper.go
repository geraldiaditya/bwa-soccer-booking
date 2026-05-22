package services

import (
	"field-service/common/utils"
	"field-service/domain/dto"
	"field-service/domain/models"
	"fmt"
	"time"
)

func toFieldScheduleResponse(schedule models.FieldSchedule) dto.FieldScheduleResponse {
	return dto.FieldScheduleResponse{
		UUID:         schedule.UUID,
		FieldName:    schedule.Field.Name,
		PricePerHour: schedule.Field.PricePerHour,
		Date:         schedule.Date.Format(time.DateOnly),
		Status:       schedule.Status.GetStatusString(),
		Time:         formatScheduleTime(schedule.Time),
		CreatedAt:    schedule.CreatedAt,
		UpdatedAt:    schedule.UpdatedAt,
	}
}

func toFieldScheduleResponseWithTime(schedule models.FieldSchedule, scheduleTime models.Time) dto.FieldScheduleResponse {
	response := toFieldScheduleResponse(schedule)
	response.Time = formatScheduleTime(scheduleTime)
	return response
}

func toFieldScheduleResponses(schedules []models.FieldSchedule) []dto.FieldScheduleResponse {
	responses := make([]dto.FieldScheduleResponse, 0, len(schedules))
	for _, schedule := range schedules {
		responses = append(responses, toFieldScheduleResponse(schedule))
	}
	return responses
}

func toFieldScheduleForBookingResponse(schedule models.FieldSchedule) dto.FieldScheduleForBookingResponse {
	pricePerHour := float64(schedule.Field.PricePerHour)
	return dto.FieldScheduleForBookingResponse{
		UUID:         schedule.UUID,
		Date:         utils.ConvertMonthName(schedule.Date.Format(time.DateOnly)),
		Time:         schedule.Time.StartTime,
		Status:       schedule.Status.GetStatusString(),
		PricePerHour: utils.RupiahFormat(&pricePerHour),
	}
}

func toFieldScheduleForBookingResponses(schedules []models.FieldSchedule) []dto.FieldScheduleForBookingResponse {
	responses := make([]dto.FieldScheduleForBookingResponse, 0, len(schedules))
	for _, schedule := range schedules {
		responses = append(responses, toFieldScheduleForBookingResponse(schedule))
	}
	return responses
}

func formatScheduleTime(scheduleTime models.Time) string {
	return fmt.Sprintf("%s - %s", scheduleTime.StartTime, scheduleTime.EndTime)
}
