package services

import (
	"context"
	errTime "field-service/constants/error/time"
	"field-service/domain/dto"
	"field-service/domain/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newSvc() (ITimeService, *mockRepositoryRegistry) {
	reg := newMockRegistry()
	return NewTimeService(reg), reg
}

func TestGetByUUID_Success(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()
	timeModel := &models.Time{UUID: id, StartTime: "08:00", EndTime: "09:00"}

	reg.timeRepo.On("FindByUUID", context.Background(), id.String()).Return(timeModel, nil)

	resp, err := svc.GetByUUID(context.Background(), id.String())

	assert.NoError(t, err)
	assert.Equal(t, id, resp.UUID)
	assert.Equal(t, "08:00", resp.StartTime)
}

func TestGetByUUID_NotFound(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()

	reg.timeRepo.On("FindByUUID", context.Background(), id.String()).Return(nil, errTime.ErrTimeNotFound)

	resp, err := svc.GetByUUID(context.Background(), id.String())

	assert.ErrorIs(t, err, errTime.ErrTimeNotFound)
	assert.Nil(t, resp)
}

func TestCreate_Success(t *testing.T) {
	svc, reg := newSvc()
	id := uuid.New()
	timeModel := &models.Time{UUID: id, StartTime: "08:00", EndTime: "09:00"}

	reg.timeRepo.On("Create", context.Background(), &models.Time{StartTime: "08:00", EndTime: "09:00"}).Return(timeModel, nil)

	resp, err := svc.Create(context.Background(), &dto.TimeRequest{StartTime: "08:00", EndTime: "09:00"})

	assert.NoError(t, err)
	assert.Equal(t, id, resp.UUID)
	assert.Equal(t, "08:00", resp.StartTime)
}
