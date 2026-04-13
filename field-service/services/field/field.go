package services

import (
	"bytes"
	"context"
	"field-service/common/gcs"
	"field-service/common/utils"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"path"
	"time"
)

type IFieldService interface {
	GetAllWithPagination(context.Context, *dto.FieldRequestParam) (*utils.PaginationResult, error)
	GetAllWithoutPagination(context.Context) ([]dto.FieldResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldResponse, error)
	Create(context.Context, *dto.FieldRequest) (*dto.FieldResponse, error)
	Update(context.Context, string, *dto.UpdateFieldRequest) (*dto.FieldResponse, error)
	Delete(context.Context, string) error
}

type FieldService struct {
	repository repositories.IRepositoryRegistry
	gsc        gcs.IGSClient
}

func NewFieldService(repository repositories.IRepositoryRegistry, gsc gcs.IGSClient) IFieldService {
	return &FieldService{repository: repository, gsc: gsc}
}

func (f *FieldService) GetAllWithPagination(
	ctx context.Context,
	param *dto.FieldRequestParam,
) (*utils.PaginationResult, error) {
	fields, total, err := f.repository.GetField().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	fieldResult := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResult = append(fieldResult, dto.FieldResponse{
			UUID:         field.UUID,
			Code:         field.Code,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		})
	}
	pagination := &utils.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  fieldResult,
	}

	response := utils.GeneratePagination(*pagination)
	return &response, nil
}

func (f *FieldService) GetAllWithoutPagination(ctx context.Context) ([]dto.FieldResponse, error) {
	fields, err := f.repository.GetField().FindAllWithoutPagination(ctx)
	if err != nil {
		return nil, err
	}

	fieldResult := make([]dto.FieldResponse, 0, len(fields))
	for _, field := range fields {
		fieldResult = append(fieldResult, dto.FieldResponse{
			UUID:         field.UUID,
			Name:         field.Name,
			PricePerHour: field.PricePerHour,
			Images:       field.Images,
		})
	}

	return fieldResult, nil
}

func (f *FieldService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldResponse, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	fieldResult := dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}
	return &fieldResult, nil
}

func (f *FieldService) Create(ctx context.Context, request *dto.FieldRequest) (*dto.FieldResponse, error) {
	imageUrl, err := f.uploadImage(ctx, request.Images)
	if err != nil {
		return nil, err
	}
	field, err := f.repository.GetField().Create(ctx, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrl,
	})
	if err != nil {
		return nil, err
	}
	response := &dto.FieldResponse{
		UUID:         field.UUID,
		Code:         field.Code,
		Name:         field.Name,
		PricePerHour: field.PricePerHour,
		Images:       field.Images,
		CreatedAt:    field.CreatedAt,
		UpdatedAt:    field.UpdatedAt,
	}
	return response, nil
}

func (f *FieldService) Update(ctx context.Context, uuidParams string, request *dto.UpdateFieldRequest) (*dto.FieldResponse, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuidParams)
	if err != nil {
		return nil, err
	}
	var imageUrl []string
	if request.Images == nil {
		imageUrl = field.Images
	} else {
		imageUrl, err = f.uploadImage(ctx, request.Images)
		if err != nil {
			return nil, err
		}
	}
	fieldResult, err := f.repository.GetField().Update(ctx, uuidParams, &models.Field{
		Code:         request.Code,
		Name:         request.Name,
		PricePerHour: request.PricePerHour,
		Images:       imageUrl,
	})
	if err != nil {
		return nil, err
	}
	uuidParsed, _ := uuid.Parse(uuidParams)
	return &dto.FieldResponse{
		UUID:         uuidParsed,
		Code:         fieldResult.Code,
		Name:         fieldResult.Name,
		PricePerHour: fieldResult.PricePerHour,
		Images:       fieldResult.Images,
		CreatedAt:    fieldResult.CreatedAt,
		UpdatedAt:    fieldResult.UpdatedAt,
	}, nil
}

func (f *FieldService) Delete(ctx context.Context, uuid string) error {
	_, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	err = f.repository.GetField().Delete(ctx, uuid)
	if err != nil {
		return err
	}
	return nil
}

func (f *FieldService) validateUpload(images []multipart.FileHeader) error {
	if images == nil || len(images) == 0 {
		return errConstant.ErrInvalidUploadFile
	}
	for _, image := range images {
		if image.Size > 5*1024*1024 {
			return errConstant.ErrSizeTooBig
		}
	}
	return nil
}

func (f *FieldService) processAndUploadImage(ctx context.Context, image multipart.FileHeader) (string, error) {
	file, err := image.Open()
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("images/%s-%s-%s", time.Now().Format("20060102150405"), image.Filename, path.Ext(image.Filename))
	url, err := f.gsc.UploadFile(ctx, filename, buffer.Bytes())

	if err != nil {
		return "", err
	}
	return url, nil
}

func (f *FieldService) uploadImage(ctx context.Context, images []multipart.FileHeader) ([]string, error) {
	err := f.validateUpload(images)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(images))
	for _, image := range images {
		url, err := f.processAndUploadImage(ctx, image)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}
