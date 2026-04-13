package services

import (
	"context"
	"field-service/domain/models"
	repoField "field-service/repositories/field"
	repoFieldSchedule "field-service/repositories/field_schedule"
	repoTime "field-service/repositories/time"

	"github.com/stretchr/testify/mock"
)

type mockTimeRepository struct {
	mock.Mock
}

func (m *mockTimeRepository) FindAll(ctx context.Context) ([]models.Time, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Time), args.Error(1)
}

func (m *mockTimeRepository) FindByUUID(ctx context.Context, uuid string) (*models.Time, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Time), args.Error(1)
}

func (m *mockTimeRepository) FindByID(ctx context.Context, id int) (*models.Time, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Time), args.Error(1)
}

func (m *mockTimeRepository) Create(ctx context.Context, req *models.Time) (*models.Time, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Time), args.Error(1)
}

type mockRepositoryRegistry struct {
	timeRepo *mockTimeRepository
}

func newMockRegistry() *mockRepositoryRegistry {
	return &mockRepositoryRegistry{timeRepo: &mockTimeRepository{}}
}

func (m *mockRepositoryRegistry) GetField() repoField.IFieldRepository {
	return nil
}

func (m *mockRepositoryRegistry) GetFieldSchedule() repoFieldSchedule.IFieldScheduleRepository {
	return nil
}

func (m *mockRepositoryRegistry) GetTime() repoTime.ITimeRepository {
	return m.timeRepo
}
