package services

import (
	"field-service/common/gcs"
	"field-service/repositories"
	fieldService "field-service/services/field"
	fieldScheduleService "field-service/services/field_schedule"
	timeService "field-service/services/time"
)

func NewServiceRegistry(repository repositories.IRepositoryRegistry, gcs gcs.IGSClient) IServiceRegistry {
	return &Registry{repository: repository, gcs: gcs}
}

type IServiceRegistry interface {
	GetField() fieldService.IFieldService
	GetFieldSchedule() fieldScheduleService.IFieldScheduleService
	GetTime() timeService.ITimeService
}

type Registry struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGSClient
}

func (r *Registry) GetField() fieldService.IFieldService {
	return fieldService.NewFieldService(r.repository, r.gcs)
}

func (r *Registry) GetFieldSchedule() fieldScheduleService.IFieldScheduleService {
	return fieldScheduleService.NewFieldScheduleService(r.repository)
}

func (r *Registry) GetTime() timeService.ITimeService {
	return timeService.NewTimeService(r.repository)
}
