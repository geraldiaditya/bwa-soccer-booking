package services

import (
	"context"
	"field-service/constants"
	errField "field-service/constants/error/field"
	errFieldSchedule "field-service/constants/error/field_schedule"
	"field-service/domain/dto"
	"field-service/domain/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newSvc() (IFieldScheduleService, *mockRepositoryRegistry) {
	reg := newMockRegistry()
	return NewFieldScheduleService(reg), reg
}

func TestGetByUUID_Success(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()
	schedule := &models.FieldSchedule{
		UUID:  id,
		Field: models.Field{Name: "Field 1", PricePerHour: 100000},
		Time:  models.Time{StartTime: "08:00", EndTime: "09:00"},
	}

	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), id.String()).Return(schedule, nil)

	resp, err := svc.GetByUUID(context.Background(), id.String())

	assert.NoError(t, err)
	assert.Equal(t, id, resp.UUID)
	assert.Equal(t, "Field 1", resp.FieldName)
}

func TestGetByUUID_NotFound(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()

	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), id.String()).Return(nil, errFieldSchedule.ErrFieldScheduleNotFound)

	resp, err := svc.GetByUUID(context.Background(), id.String())

	assert.ErrorIs(t, err, errFieldSchedule.ErrFieldScheduleNotFound)
	assert.Nil(t, resp)
}

func TestDelete_Success(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()
	schedule := &models.FieldSchedule{UUID: id}

	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), id.String()).Return(schedule, nil)
	reg.fieldScheduleRepo.On("Delete", context.Background(), id.String()).Return(nil)

	err := svc.Delete(context.Background(), id.String())

	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()

	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), id.String()).Return(nil, errFieldSchedule.ErrFieldScheduleNotFound)

	err := svc.Delete(context.Background(), id.String())

	assert.ErrorIs(t, err, errFieldSchedule.ErrFieldScheduleNotFound)
}

func TestGenerateScheduleForOneMonth_DocumentsBusinessScenarios(t *testing.T) {
	ctx := context.Background()
	fieldID := uuid.New().String()
	field := &models.Field{ID: 7, UUID: uuid.MustParse(fieldID)}
	times := []models.Time{
		{ID: 1, UUID: uuid.New()},
		{ID: 2, UUID: uuid.New()},
	}

	tests := []struct {
		name      string
		arrange   func(*mockRepositoryRegistry)
		wantError error
	}{
		{
			name: "successful generation creates thirty days for every time slot after one batch conflict lookup",
			arrange: func(reg *mockRepositoryRegistry) {
				reg.fieldRepo.On("FindByUUID", ctx, fieldID).Return(field, nil)
				reg.timeRepo.On("FindAll", ctx).Return(times, nil)
				reg.fieldScheduleRepo.On(
					"FindAllByFieldIDAndDateRange",
					ctx,
					int(field.ID),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return([]models.FieldSchedule{}, nil).Once()
				reg.fieldScheduleRepo.On("Create", ctx, mock.MatchedBy(func(schedules []models.FieldSchedule) bool {
					if len(schedules) != 30*len(times) {
						return false
					}
					for _, schedule := range schedules {
						if schedule.FieldID != field.ID || schedule.Status != constants.Available {
							return false
						}
					}
					return true
				})).Return(nil)
			},
		},
		{
			name: "field not found returns field error before loading time slots",
			arrange: func(reg *mockRepositoryRegistry) {
				reg.fieldRepo.On("FindByUUID", ctx, fieldID).Return(nil, errField.ErrFieldNotFound)
			},
			wantError: errField.ErrFieldNotFound,
		},
		{
			name: "existing schedule conflict returns conflict error without creating duplicates",
			arrange: func(reg *mockRepositoryRegistry) {
				reg.fieldRepo.On("FindByUUID", ctx, fieldID).Return(field, nil)
				reg.timeRepo.On("FindAll", ctx).Return(times, nil)
				reg.fieldScheduleRepo.On(
					"FindAllByFieldIDAndDateRange",
					ctx,
					int(field.ID),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return([]models.FieldSchedule{{ID: 1, FieldID: field.ID}}, nil).Once()
			},
			wantError: errFieldSchedule.ErrFieldScheduleIsExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, reg := newSvc()
			tt.arrange(reg)

			err := svc.GenerateScheduleForOneMonth(ctx, &dto.GenerateFieldScheduleForOneMonthRequest{FieldID: fieldID})

			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
			reg.fieldRepo.AssertExpectations(t)
			reg.timeRepo.AssertExpectations(t)
			reg.fieldScheduleRepo.AssertExpectations(t)
		})
	}
}
