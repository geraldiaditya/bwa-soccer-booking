package services

import (
	"context"
	"errors"
	"field-service/constants"
	errorConstants "field-service/constants/error"
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

func TestUpdateStatus_ReturnsErrorInsideTransactionWhenSecondUpdateFails(t *testing.T) {
	svc, reg := newSvc()
	firstID := uuid.New().String()
	secondID := uuid.New().String()
	expectedErr := errors.New("update status failed")
	request := &dto.UpdateStatusFieldScheduleRequest{
		FieldScheduleIDs: []string{firstID, secondID},
	}

	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), firstID).
		Return(&models.FieldSchedule{UUID: uuid.MustParse(firstID)}, nil)
	reg.fieldScheduleRepo.On("UpdateStatus", context.Background(), constants.Booked, firstID).
		Return(nil)
	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), secondID).
		Return(&models.FieldSchedule{UUID: uuid.MustParse(secondID)}, nil)
	reg.fieldScheduleRepo.On("UpdateStatus", context.Background(), constants.Booked, secondID).
		Return(expectedErr)

	err := svc.UpdateStatus(context.Background(), request)

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 1, reg.transactionCalls)
	reg.fieldScheduleRepo.AssertExpectations(t)
}

func TestUpdateStatus_SucceedsInsideTransactionWhenAllSchedulesAreUpdated(t *testing.T) {
	svc, reg := newSvc()
	firstID := uuid.New().String()
	secondID := uuid.New().String()
	request := &dto.UpdateStatusFieldScheduleRequest{
		FieldScheduleIDs: []string{firstID, secondID},
	}

	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), firstID).
		Return(&models.FieldSchedule{UUID: uuid.MustParse(firstID)}, nil)
	reg.fieldScheduleRepo.On("UpdateStatus", context.Background(), constants.Booked, firstID).
		Return(nil)
	reg.fieldScheduleRepo.On("FindByUUID", context.Background(), secondID).
		Return(&models.FieldSchedule{UUID: uuid.MustParse(secondID)}, nil)
	reg.fieldScheduleRepo.On("UpdateStatus", context.Background(), constants.Booked, secondID).
		Return(nil)

	err := svc.UpdateStatus(context.Background(), request)

	assert.NoError(t, err)
	assert.Equal(t, 1, reg.transactionCalls)
	reg.fieldScheduleRepo.AssertExpectations(t)
}

func TestCreate_SucceedsInsideTransaction(t *testing.T) {
	svc, reg := newSvc()
	fieldID := uuid.New().String()
	timeID := uuid.New().String()
	field := &models.Field{ID: 1, UUID: uuid.MustParse(fieldID)}
	scheduleTime := &models.Time{ID: 2, UUID: uuid.MustParse(timeID)}
	request := &dto.FieldScheduleRequest{
		FieldID: fieldID,
		Date:    "2026-06-01",
		TimeIDs: []string{timeID},
	}

	reg.fieldRepo.On("FindByUUID", context.Background(), fieldID).Return(field, nil)
	reg.timeRepo.On("FindByUUID", context.Background(), timeID).Return(scheduleTime, nil)
	reg.fieldScheduleRepo.On("FindByDateAndTimeId", context.Background(), request.Date, int(scheduleTime.ID), int(field.ID)).
		Return(nil, nil)
	reg.fieldScheduleRepo.On("Create", context.Background(), mock.MatchedBy(func(schedules []models.FieldSchedule) bool {
		return len(schedules) == 1 &&
			schedules[0].FieldID == field.ID &&
			schedules[0].TimeID == scheduleTime.ID &&
			schedules[0].Status == constants.Available
	})).Return(nil)

	err := svc.Create(context.Background(), request)

	assert.NoError(t, err)
	assert.Equal(t, 1, reg.transactionCalls)
	reg.fieldRepo.AssertExpectations(t)
	reg.timeRepo.AssertExpectations(t)
	reg.fieldScheduleRepo.AssertExpectations(t)
}

func TestCreate_ReturnsValidationErrorWhenDateCannotBeParsed(t *testing.T) {
	svc, reg := newSvc()
	fieldID := uuid.New().String()
	field := &models.Field{ID: 1, UUID: uuid.MustParse(fieldID)}
	request := &dto.FieldScheduleRequest{
		FieldID: fieldID,
		Date:    "not-a-date",
		TimeIDs: []string{uuid.New().String()},
	}

	reg.fieldRepo.On("FindByUUID", context.Background(), fieldID).Return(field, nil)

	err := svc.Create(context.Background(), request)

	assert.ErrorIs(t, err, errorConstants.ErrRequestValidation)
	assert.Equal(t, 1, reg.transactionCalls)
	reg.fieldRepo.AssertExpectations(t)
}
