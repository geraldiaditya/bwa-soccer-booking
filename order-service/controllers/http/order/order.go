package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	errValidation "order-service/common/error"
	"order-service/common/response"
	"order-service/domain/dto"
	"order-service/services"
)

func NewOrderController(services services.IServiceRegistry) IOrderController {
	return &OrderController{
		service: services,
	}
}

type IOrderController interface {
	GetAllWithPagination(ctx *gin.Context)
	GetByUUID(ctx *gin.Context)
	GetOrderByUserId(ctx *gin.Context)
	Create(ctx *gin.Context)
}
type OrderController struct {
	service services.IServiceRegistry
}

func (o *OrderController) GetAllWithPagination(ctx *gin.Context) {
	var params dto.OrderRequestParam
	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Err:  err,
				Gin:  ctx,
			})
		return
	}
	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(
			response.ParamHTTPResp{
				Code:    http.StatusUnprocessableEntity,
				Err:     err,
				Message: &errMessage,
				Gin:     ctx,
				Data:    errResponse,
			})
		return
	}
	result, err := o.service.GetOrder().GetAllWithPagination(ctx, &params)
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Err:  err,
				Gin:  ctx,
			})
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Gin:  ctx,
			Data: result,
		})
}

func (o *OrderController) GetByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	result, err := o.service.GetOrder().GetByUUID(ctx, uuid)
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Err:  err,
				Gin:  ctx,
			})
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Gin:  ctx,
			Data: result,
		})
}

func (o *OrderController) GetOrderByUserId(ctx *gin.Context) {
	result, err := o.service.GetOrder().GetOrderByUserID(ctx.Request.Context())
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Gin:  ctx,
				Err:  err,
			})
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Gin:  ctx,
			Data: result,
		})
}

func (o *OrderController) Create(ctx *gin.Context) {
	var (
		request  dto.OrderRequest
		rContext = ctx.Request.Context()
	)
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Gin:  ctx,
				Err:  err,
			})
		return
	}
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(
			response.ParamHTTPResp{
				Code:    http.StatusUnprocessableEntity,
				Err:     err,
				Message: &errMessage,
				Gin:     ctx,
				Data:    errResponse,
			})
		return
	}
	result, err := o.service.GetOrder().Create(rContext, &request)
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: http.StatusBadRequest,
				Gin:  ctx,
				Err:  err,
			})
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Gin:  ctx,
			Data: result,
		})
}
