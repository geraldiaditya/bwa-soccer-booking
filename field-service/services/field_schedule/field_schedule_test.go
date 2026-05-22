package services

import (
	"context"
	"field-service/constants"
	errFieldSchedule "field-service/constants/error/field_schedule"
	errTime "field-service/constants/error/time"
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

func TestCreate_DocumentsScheduleCreationBusinessRules(t *testing.T) {
	ctx := context.Background()
	fieldID := uuid.New().String()
	firstTimeID := uuid.New().String()
	secondTimeID := uuid.New().String()
	field := &models.Field{ID: 3, UUID: uuid.MustParse(fieldID)}
	firstTime := &models.Time{ID: 10, UUID: uuid.MustParse(firstTimeID), StartTime: "08:00", EndTime: "09:00"}
	secondTime := &models.Time{ID: 11, UUID: uuid.MustParse(secondTimeID), StartTime: "09:00", EndTime: "10:00"}
	request := &dto.FieldScheduleRequest{
		FieldID: fieldID,
		Date:    "2026-06-01",
		TimeIDs: []string{firstTimeID, secondTimeID},
	}

	tests := []struct {
		name      string
		arrange   func(*mockRepositoryRegistry)
		wantError error
	}{
		{
			name: "successful creation writes one available schedule per selected time slot",
			arrange: func(reg *mockRepositoryRegistry) {
				reg.fieldRepo.On("FindByUUID", ctx, fieldID).Return(field, nil)
				reg.timeRepo.On("FindByUUID", ctx, firstTimeID).Return(firstTime, nil)
				reg.fieldScheduleRepo.On("FindByDateAndTimeId", ctx, request.Date, int(firstTime.ID), int(field.ID)).Return(nil, nil)
				reg.timeRepo.On("FindByUUID", ctx, secondTimeID).Return(secondTime, nil)
				reg.fieldScheduleRepo.On("FindByDateAndTimeId", ctx, request.Date, int(secondTime.ID), int(field.ID)).Return(nil, nil)
				reg.fieldScheduleRepo.On("Create", ctx, mock.MatchedBy(func(schedules []models.FieldSchedule) bool {
					return len(schedules) == 2 &&
						schedules[0].Status == constants.Available &&
						schedules[1].Status == constants.Available &&
						schedules[0].FieldID == field.ID &&
						schedules[1].FieldID == field.ID
				})).Return(nil)
			},
		},
		{
			name: "missing time slot stops before creating schedules",
			arrange: func(reg *mockRepositoryRegistry) {
				reg.fieldRepo.On("FindByUUID", ctx, fieldID).Return(field, nil)
				reg.timeRepo.On("FindByUUID", ctx, firstTimeID).Return(nil, errTime.ErrTimeNotFound)
			},
			wantError: errTime.ErrTimeNotFound,
		},
		{
			name: "existing schedule conflict stops before creating duplicate schedules",
			arrange: func(reg *mockRepositoryRegistry) {
				reg.fieldRepo.On("FindByUUID", ctx, fieldID).Return(field, nil)
				reg.timeRepo.On("FindByUUID", ctx, firstTimeID).Return(firstTime, nil)
				reg.fieldScheduleRepo.On("FindByDateAndTimeId", ctx, request.Date, int(firstTime.ID), int(field.ID)).
					Return(&models.FieldSchedule{ID: 1}, nil)
			},
			wantError: errFieldSchedule.ErrFieldScheduleIsExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, reg := newSvc()
			tt.arrange(reg)

			err := svc.Create(ctx, request)

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
