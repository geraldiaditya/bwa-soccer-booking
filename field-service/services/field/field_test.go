package services

import (
	"context"
	errConstant "field-service/constants/error"
	errField "field-service/constants/error/field"
	"field-service/domain/dto"
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

func TestGetAllWithPagination_ReturnsBusinessDataWithPaginationMeta(t *testing.T) {
	svc, reg, _ := newSvc()
	param := &dto.FieldRequestParam{Page: 2, Limit: 1}
	fields := []models.Field{
		{UUID: uuid.New(), Name: "Field 2", Code: "F2", PricePerHour: 150000, Images: []string{"field-2.jpg"}},
	}

	reg.fieldRepo.On("FindAllWithPagination", context.Background(), param).Return(fields, 3, nil)

	resp, err := svc.GetAllWithPagination(context.Background(), param)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), resp.TotalData)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 1, resp.Limit)
	assert.Len(t, resp.Data, 1)
}

func TestGetAllWithoutPagination_ReturnsFieldSummaries(t *testing.T) {
	svc, reg, _ := newSvc()
	fields := []models.Field{
		{UUID: uuid.New(), Name: "Field 1", Code: "F1", PricePerHour: 100000, Images: []string{"field-1.jpg"}},
	}

	reg.fieldRepo.On("FindAllWithoutPagination", context.Background()).Return(fields, nil)

	resp, err := svc.GetAllWithoutPagination(context.Background())

	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Field 1", resp[0].Name)
	assert.Empty(t, resp[0].Code)
}

func TestCreate_RejectsMissingImagesBeforeWritingField(t *testing.T) {
	svc, _, _ := newSvc()
	request := &dto.FieldRequest{Name: "Field 1", Code: "F1", PricePerHour: 100000}

	resp, err := svc.Create(context.Background(), request)

	assert.ErrorIs(t, err, errConstant.ErrInvalidUploadFile)
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

func TestUpdate_UsesExistingFieldUUID(t *testing.T) {
	svc, reg, _ := newSvc()
	id := uuid.New()
	existingField := &models.Field{UUID: id, Images: []string{"old-image"}}
	updatedField := &models.Field{Name: "Updated Field", Code: "UF", PricePerHour: 120000, Images: []string{"old-image"}}
	request := &dto.UpdateFieldRequest{Name: "Updated Field", Code: "UF", PricePerHour: 120000}

	reg.fieldRepo.On("FindByUUID", context.Background(), id.String()).Return(existingField, nil)
	reg.fieldRepo.On("Update", context.Background(), id.String(), updatedField).Return(updatedField, nil)

	resp, err := svc.Update(context.Background(), id.String(), request)

	assert.NoError(t, err)
	assert.Equal(t, id, resp.UUID)
	assert.Equal(t, "Updated Field", resp.Name)
}

func TestDelete_NotFound(t *testing.T) {
	svc, reg, _ := newSvc()
	id := uuid.New()

	reg.fieldRepo.On("FindByUUID", context.Background(), id.String()).Return(nil, errField.ErrFieldNotFound)

	err := svc.Delete(context.Background(), id.String())

	assert.ErrorIs(t, err, errField.ErrFieldNotFound)
}
