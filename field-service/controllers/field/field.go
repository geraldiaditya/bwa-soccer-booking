package controllers

import (
	errWrap "field-service/common/error"
	"field-service/common/response"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/services"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IFieldController interface {
	GetAllWithPagination(*gin.Context)
	GetAllWithoutPagination(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
}

func NewFieldController(service services.IServiceRegistry) IFieldController {
	return &FieldController{service: service}
}

type FieldController struct {
	service services.IServiceRegistry
}

func (f *FieldController) validateFiles(files []multipart.FileHeader) error {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	for _, file := range files {
		if file.Size > 5*1024*1024 {
			return errConstant.ErrSizeTooBig
		}

		f, err := file.Open()
		if err != nil {
			return err
		}
		defer f.Close()

		buffer := make([]byte, 512)
		_, err = f.Read(buffer)
		if err != nil {
			return err
		}

		contentType := http.DetectContentType(buffer)
		if !allowedTypes[contentType] {
			return errConstant.ErrInvalidUploadFile
		}
	}
	return nil
}

// GetAllWithPagination handles getting all fields with pagination.
// @Summary Get all fields with pagination
// @Tags Field
// @Accept json
// @Produce json
// @Param page query int true "Page number" default(1)
// @Param limit query int true "Limit per page" default(10)
// @Param sortColumn query string false "Sort column"
// @Param sortOrder query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /fields [get]
func (f *FieldController) GetAllWithPagination(context *gin.Context) {
	var params dto.FieldRequestParam
	err := context.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	validate := validator.New()
	err = validate.Struct(params)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errWrap.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    errConstant.ErrStatusCode(err),
			Err:     err,
			Gin:     context,
			Message: &errMessage,
			Data:    errResponse,
		})
		return
	}
	result, err := f.service.GetField().GetAllWithPagination(context, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

// GetAllWithoutPagination handles getting all fields without pagination.
// @Summary Get all fields without pagination
// @Tags Field
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /fields/all [get]
func (f *FieldController) GetAllWithoutPagination(context *gin.Context) {
	result, err := f.service.GetField().GetAllWithoutPagination(context)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

// GetByUUID handles getting a field by its UUID.
// @Summary Get field by UUID
// @Tags Field
// @Accept json
// @Produce json
// @Param uuid path string true "Field UUID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /fields/{uuid} [get]
func (f *FieldController) GetByUUID(context *gin.Context) {
	result, err := f.service.GetField().GetByUUID(context, context.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

// Create handles creating a new field.
// @Summary Create a new field
// @Tags Field
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Field Name"
// @Param code formData string true "Field Code"
// @Param pricePerHour formData int true "Price Per Hour"
// @Param images formData file true "Field Images"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /fields [post]
func (f *FieldController) Create(context *gin.Context) {
	var request dto.FieldRequest
	err := context.ShouldBindWith(&request, binding.FormMultipart)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errWrap.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    errConstant.ErrStatusCode(err),
			Err:     err,
			Gin:     context,
			Message: &errMessage,
			Data:    errResponse,
		})
		return
	}

	if err := f.validateFiles(request.Images); err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}

	result, err := f.service.GetField().Create(context, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

// Update handles updating an existing field.
// @Summary Update a field
// @Tags Field
// @Accept multipart/form-data
// @Produce json
// @Param uuid path string true "Field UUID"
// @Param name formData string true "Field Name"
// @Param code formData string true "Field Code"
// @Param pricePerHour formData int true "Price Per Hour"
// @Param images formData file false "Field Images"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /fields/{uuid} [put]
func (f *FieldController) Update(context *gin.Context) {
	var request dto.UpdateFieldRequest
	err := context.ShouldBindWith(&request, binding.FormMultipart)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	validate := validator.New()
	err = validate.Struct(request)
	if err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errWrap.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    errConstant.ErrStatusCode(err),
			Err:     err,
			Gin:     context,
			Message: &errMessage,
			Data:    errResponse,
		})
		return
	}

	if request.Images != nil {
		if err := f.validateFiles(request.Images); err != nil {
			response.HttpResponse(response.ParamHTTPResp{
				Code: errConstant.ErrStatusCode(err),
				Err:  err,
				Gin:  context,
			})
			return
		}
	}

	result, err := f.service.GetField().Update(context, context.Param("uuid"), &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  context,
	})
}

// Delete handles deleting a field.
// @Summary Delete a field
// @Tags Field
// @Accept json
// @Produce json
// @Param uuid path string true "Field UUID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /fields/{uuid} [delete]
func (f *FieldController) Delete(context *gin.Context) {
	err := f.service.GetField().Delete(context, context.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  context,
	})
}
