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

// GetAll handles getting all time slots.
// @Summary Get all time slots
// @Tags Time
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /times [get]
func (t *TimeController) GetAll(context *gin.Context) {
	result, err := t.service.GetTime().GetAll(context)
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
		Data: result,
	})
}

// GetByUUID handles getting a time slot by its UUID.
// @Summary Get time slot by UUID
// @Tags Time
// @Accept json
// @Produce json
// @Param uuid path string true "Time UUID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /times/{uuid} [get]
func (t *TimeController) GetByUUID(context *gin.Context) {
	result, err := t.service.GetTime().GetByUUID(context, context.Param("uuid"))
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

// Create handles creating a new time slot.
// @Summary Create a time slot
// @Tags Time
// @Accept json
// @Produce json
// @Param body body dto.TimeRequest true "Time Request"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /times [post]
func (t *TimeController) Create(context *gin.Context) {
	var request dto.TimeRequest
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
	result, err := t.service.GetTime().Create(context, &request)
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
