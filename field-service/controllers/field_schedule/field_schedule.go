package controllers

import (
	errWrap "field-service/common/error"
	"field-service/common/response"
	errConstant "field-service/constants/error"
	"field-service/domain/dto"
	"field-service/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type IFieldScheduleController interface {
	GetAllWithPagination(*gin.Context)
	GetAllByFieldIdAndDate(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
	Update(*gin.Context)
	UpdateStatus(*gin.Context)
	Delete(*gin.Context)
	GenerateScheduleForOneMonth(*gin.Context)
}

type FieldScheduleController struct {
	service services.IServiceRegistry
}

func NewFieldScheduleController(service services.IServiceRegistry) IFieldScheduleController {
	return &FieldScheduleController{service: service}
}

// GetAllWithPagination handles getting all field schedules with pagination.
// @Summary Get all field schedules with pagination
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param page query int true "Page number" default(1)
// @Param limit query int true "Limit per page" default(10)
// @Param sortColumn query string false "Sort column"
// @Param sortOrder query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /field-schedules [get]
func (f *FieldScheduleController) GetAllWithPagination(context *gin.Context) {
	var params dto.FieldScheduleRequestParam
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
	result, err := f.service.GetFieldSchedule().GetAllWithPagination(context, &params)
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

// GetAllByFieldIdAndDate handles getting field schedules for a specific field and date.
// @Summary Get field schedules by field ID and date
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param uuid path string true "Field UUID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /field-schedules/field/{uuid} [get]
func (f *FieldScheduleController) GetAllByFieldIdAndDate(context *gin.Context) {
	var params dto.FieldScheduleByFieldIDAndDateRequestParam
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
	result, err := f.service.GetFieldSchedule().GetAllByFieldIdAndDate(context, context.Param("uuid"), params.Date)
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

// GetByUUID handles getting a field schedule by its UUID.
// @Summary Get field schedule by UUID
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param uuid path string true "Schedule UUID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /field-schedules/{uuid} [get]
func (f *FieldScheduleController) GetByUUID(context *gin.Context) {
	result, err := f.service.GetFieldSchedule().GetByUUID(context, context.Param("uuid"))
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

// Create handles creating a new field schedule.
// @Summary Create a field schedule
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param body body dto.FieldScheduleRequest true "Schedule Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /field-schedules [post]
func (f *FieldScheduleController) Create(context *gin.Context) {
	var params dto.FieldScheduleRequest
	err := context.ShouldBindJSON(&params)
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
	err = f.service.GetFieldSchedule().Create(context, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  context,
	})

}

// Update handles updating an existing field schedule.
// @Summary Update a field schedule
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param uuid path string true "Schedule UUID"
// @Param body body dto.UpdateFieldScheduleRequest true "Update Request"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /field-schedules/{uuid} [put]
func (f *FieldScheduleController) Update(context *gin.Context) {
	var params dto.UpdateFieldScheduleRequest
	err := context.ShouldBindJSON(&params)
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
	result, err := f.service.GetFieldSchedule().Update(context, context.Param("uuid"), &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  context,
		Data: result,
	})
}

// UpdateStatus handles updating the status of field schedules.
// @Summary Update schedule status
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param body body dto.UpdateStatusFieldScheduleRequest true "Status Update Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /field-schedules/status [put]
func (f *FieldScheduleController) UpdateStatus(context *gin.Context) {
	var request dto.UpdateStatusFieldScheduleRequest
	err := context.ShouldBindJSON(&request)
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
	err = f.service.GetFieldSchedule().UpdateStatus(context, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  context,
	})
}

// Delete handles deleting a field schedule.
// @Summary Delete a field schedule
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param uuid path string true "Schedule UUID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /field-schedules/{uuid} [delete]
func (f *FieldScheduleController) Delete(context *gin.Context) {
	err := f.service.GetFieldSchedule().Delete(context, context.Param("uuid"))
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

// GenerateScheduleForOneMonth handles generating schedules for a month.
// @Summary Generate schedules for one month
// @Tags FieldSchedule
// @Accept json
// @Produce json
// @Param body body dto.GenerateFieldScheduleForOneMonthRequest true "Generate Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /field-schedules/generate [post]
func (f *FieldScheduleController) GenerateScheduleForOneMonth(context *gin.Context) {
	var params dto.GenerateFieldScheduleForOneMonthRequest
	err := context.ShouldBindJSON(&params)
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
	err = f.service.GetFieldSchedule().GenerateScheduleForOneMonth(context, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: errConstant.ErrStatusCode(err),
			Err:  err,
			Gin:  context,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  context,
	})
}
