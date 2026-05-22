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
	return toTimeResponses(times), nil
}

func (t *TimeService) GetByUUID(ctx context.Context, uuid string) (*dto.TimeResponse, error) {
	time, err := t.repository.GetTime().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	response := toTimeResponse(*time)
	return &response, nil
}

func (t *TimeService) Create(ctx context.Context, request *dto.TimeRequest) (*dto.TimeResponse, error) {

	time, err := t.repository.GetTime().Create(ctx, &models.Time{
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
	})
	if err != nil {
		return nil, err
	}
	response := toTimeResponse(*time)
	return &response, nil
}
