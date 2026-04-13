package services

import (
	"context"
	errField "field-service/constants/error/field"
	"field-service/domain/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newSvc() (IFieldService, *mockRepositoryRegistry, *mockGCSClient) {
	reg := newMockRegistry()
	gcs := &mockGCSClient{}
	return NewFieldService(reg, gcs), reg, gcs
}

func TestGetByUUID_Success(t *testing.T) {
	svc, reg, _ := newSvc()
	id := uuid.New()
	field := &models.Field{UUID: id, Name: "Field 1", Code: "F1"}

	reg.fieldRepo.On("FindByUUID", context.Background(), id.String()).Return(field, nil)

	resp, err := svc.GetByUUID(context.Background(), id.String())

	assert.NoError(t, err)
	assert.Equal(t, id, resp.UUID)
	assert.Equal(t, "Field 1", resp.Name)
}

func TestGetByUUID_NotFound(t *testing.T) {
	svc, reg, _ := newSvc()
	id := uuid.New()

	reg.fieldRepo.On("FindByUUID", context.Background(), id.String()).Return(nil, errField.ErrFieldNotFound)

	resp, err := svc.GetByUUID(context.Background(), id.String())

	assert.ErrorIs(t, err, errField.ErrFieldNotFound)
	assert.Nil(t, resp)
}

func TestDelete_Success(t *testing.T) {
	svc, reg, _ := newSvc()
	id := uuid.New()
	field := &models.Field{UUID: id}

	reg.fieldRepo.On("FindByUUID", context.Background(), id.String()).Return(field, nil)
	reg.fieldRepo.On("Delete", context.Background(), id.String()).Return(nil)

	err := svc.Delete(context.Background(), id.String())

	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	svc, reg, _ := newSvc()
	id := uuid.New()

	reg.fieldRepo.On("FindByUUID", context.Background(), id.String()).Return(nil, errField.ErrFieldNotFound)

	err := svc.Delete(context.Background(), id.String())

	assert.ErrorIs(t, err, errField.ErrFieldNotFound)
}
