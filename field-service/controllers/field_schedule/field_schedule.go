package controllers

import (
	errWrap "field-service/common/error"
	"field-service/common/response"
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

func (f *FieldScheduleController) GetAllWithPagination(context *gin.Context) {
	var params dto.FieldScheduleRequestParam
	err := context.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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
			Code:    http.StatusBadRequest,
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
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) GetAllByFieldIdAndDate(context *gin.Context) {
	var params dto.FieldScheduleByFieldIDAndDateRequestParam
	err := context.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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
			Code:    http.StatusBadRequest,
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
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) GetByUUID(context *gin.Context) {
	result, err := f.service.GetFieldSchedule().GetByUUID(context, context.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) Create(context *gin.Context) {
	var params dto.FieldScheduleRequest
	err := context.ShouldBindJSON(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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
			Code:    http.StatusBadRequest,
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
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) Update(context *gin.Context) {
	var params dto.UpdateFieldScheduleRequest
	err := context.ShouldBindJSON(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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
			Code:    http.StatusBadRequest,
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
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) UpdateStatus(context *gin.Context) {
	var request dto.UpdateStatusFieldScheduleRequest
	err := context.ShouldBindJSON(&request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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
			Code:    http.StatusBadRequest,
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
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) Delete(context *gin.Context) {
	err := f.service.GetFieldSchedule().Delete(context, context.Param("uuid"))
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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

func (f *FieldScheduleController) GenerateScheduleForOneMonth(context *gin.Context) {
	var params dto.GenerateFieldScheduleForOneMonthRequest
	err := context.ShouldBindJSON(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
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
			Code:    http.StatusBadRequest,
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
			Code: http.StatusBadRequest,
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
