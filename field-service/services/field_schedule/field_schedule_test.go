package services

import (
	"context"
	errFieldSchedule "field-service/constants/error/field_schedule"
	"field-service/domain/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newSvc() (IFieldScheduleService, *mockRepositoryRegistry) {
	reg := newMockRegistry()
	return NewFieldScheduleService(reg), reg
}

func TestGetByUUID_Success(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()
	schedule := &models.FieldSchedule{
		UUID: id,
		Field: models.Field{Name: "Field 1", PricePerHour: 100000},
		Time: models.Time{StartTime: "08:00", EndTime: "09:00"},
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
