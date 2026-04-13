package services

import (
	"context"
	"field-service/constants"
	"field-service/domain/dto"
	"field-service/domain/models"
	repoField "field-service/repositories/field"
	repoFieldSchedule "field-service/repositories/field_schedule"
	repoTime "field-service/repositories/time"

	"github.com/stretchr/testify/mock"
)

type mockFieldRepository struct {
	mock.Mock
}

func (m *mockFieldRepository) FindAllWithPagination(ctx context.Context, param *dto.FieldRequestParam) ([]models.Field, int64, error) {
	args := m.Called(ctx, param)
	return args.Get(0).([]models.Field), int64(args.Int(1)), args.Error(2)
}

func (m *mockFieldRepository) FindAllWithoutPagination(ctx context.Context) ([]models.Field, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Field), args.Error(1)
}

func (m *mockFieldRepository) FindByUUID(ctx context.Context, uuid string) (*models.Field, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *mockFieldRepository) Create(ctx context.Context, req *models.Field) (*models.Field, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *mockFieldRepository) Update(ctx context.Context, uuid string, req *models.Field) (*models.Field, error) {
	args := m.Called(ctx, uuid, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Field), args.Error(1)
}

func (m *mockFieldRepository) Delete(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

type mockFieldScheduleRepository struct {
	mock.Mock
}

func (m *mockFieldScheduleRepository) FindAllWithPagination(ctx context.Context, param *dto.FieldScheduleRequestParam) ([]models.FieldSchedule, int64, error) {
	args := m.Called(ctx, param)
	return args.Get(0).([]models.FieldSchedule), int64(args.Int(1)), args.Error(2)
}

func (m *mockFieldScheduleRepository) FindAllByFieldIdAndDate(ctx context.Context, fieldId int, date string) ([]models.FieldSchedule, error) {
	args := m.Called(ctx, fieldId, date)
	return args.Get(0).([]models.FieldSchedule), args.Error(1)
}

func (m *mockFieldScheduleRepository) FindByUUID(ctx context.Context, uuid string) (*models.FieldSchedule, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FieldSchedule), args.Error(1)
}

func (m *mockFieldScheduleRepository) FindByDateAndTimeId(ctx context.Context, date string, timeId int, fieldId int) (*models.FieldSchedule, error) {
	args := m.Called(ctx, date, timeId, fieldId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FieldSchedule), args.Error(1)
}

func (m *mockFieldScheduleRepository) Create(ctx context.Context, req []models.FieldSchedule) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockFieldScheduleRepository) Update(ctx context.Context, uuid string, req *models.FieldSchedule) (*models.FieldSchedule, error) {
	args := m.Called(ctx, uuid, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FieldSchedule), args.Error(1)
}

func (m *mockFieldScheduleRepository) UpdateStatus(ctx context.Context, status constants.FieldScheduleStatus, uuid string) error {
	args := m.Called(ctx, status, uuid)
	return args.Error(0)
}

func (m *mockFieldScheduleRepository) Delete(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

type mockRepositoryRegistry struct {
	fieldRepo         *mockFieldRepository
	fieldScheduleRepo *mockFieldScheduleRepository
}

func newMockRegistry() *mockRepositoryRegistry {
	return &mockRepositoryRegistry{
		fieldRepo:         &mockFieldRepository{},
		fieldScheduleRepo: &mockFieldScheduleRepository{},
	}
}

func (m *mockRepositoryRegistry) GetField() repoField.IFieldRepository {
	return m.fieldRepo
}

func (m *mockRepositoryRegistry) GetFieldSchedule() repoFieldSchedule.IFieldScheduleRepository {
	return m.fieldScheduleRepo
}

func (m *mockRepositoryRegistry) GetTime() repoTime.ITimeRepository {
	return nil
}
