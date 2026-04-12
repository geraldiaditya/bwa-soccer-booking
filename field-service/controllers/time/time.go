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

type ITimeController interface {
	GetAll(*gin.Context)
	GetByUUID(*gin.Context)
	Create(*gin.Context)
}

func NewTimeController(service services.IServiceRegistry) ITimeController {
	return &TimeController{service: service}
}

type TimeController struct {
	service services.IServiceRegistry
}

func (t *TimeController) GetAll(context *gin.Context) {
	result, err := t.service.GetTime().GetAll(context)
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
		Data: result,
	})
}

func (t *TimeController) GetByUUID(context *gin.Context) {
	result, err := t.service.GetTime().GetByUUID(context, context.Param("uuid"))
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

func (t *TimeController) Create(context *gin.Context) {
	var request dto.TimeRequest
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
	result, err := t.service.GetTime().Create(context, &request)

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Gin:  context,
		Data: result,
	})
}
