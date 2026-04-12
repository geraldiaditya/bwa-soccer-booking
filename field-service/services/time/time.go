package services

import (
	"context"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
)

func NewTimeService(repository repositories.IRepositoryRegistry) ITimeService {
	return &TimeService{repository: repository}
}

type ITimeService interface {
	GetAll(context.Context) ([]dto.TimeResponse, error)
	GetByUUID(context.Context, string) (*dto.TimeResponse, error)
	Create(context.Context, *dto.TimeRequest) (*dto.TimeResponse, error)
}

type TimeService struct {
	repository repositories.IRepositoryRegistry
}

func (t *TimeService) GetAll(ctx context.Context) ([]dto.TimeResponse, error) {
	times, err := t.repository.GetTime().FindAll(ctx)
	if err != nil {
		return nil, err
	}
	timeResult := make([]dto.TimeResponse, 0, len(times))
	for _, time := range times {
		timeResult = append(timeResult, dto.TimeResponse{
			UUID:      time.UUID,
			StartTime: time.StartTime,
			EndTime:   time.EndTime,
			CreatedAt: time.CreatedAt,
			UpdatedAt: time.UpdatedAt,
		})
	}
	return timeResult, nil
}

func (t *TimeService) GetByUUID(ctx context.Context, uuid string) (*dto.TimeResponse, error) {
	time, err := t.repository.GetTime().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	timeResult := dto.TimeResponse{
		UUID:      time.UUID,
		StartTime: time.StartTime,
		EndTime:   time.EndTime,
		CreatedAt: time.CreatedAt,
		UpdatedAt: time.UpdatedAt,
	}
	return &timeResult, nil
}

func (t *TimeService) Create(ctx context.Context, request *dto.TimeRequest) (*dto.TimeResponse, error) {

	time, err := t.repository.GetTime().Create(ctx, &models.Time{
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
	})
	if err != nil {
		return nil, err
	}
	response := dto.TimeResponse{
		UUID:      time.UUID,
		StartTime: time.StartTime,
		EndTime:   time.EndTime,
		CreatedAt: time.CreatedAt,
		UpdatedAt: time.UpdatedAt,
	}
	return &response, nil
}
