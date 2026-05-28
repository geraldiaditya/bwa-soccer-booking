package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"payment-service/common/response"
	"payment-service/domain/dto"
	"payment-service/services"
)

type IPaymentController interface {
	GetAllWithPagination(ctx *gin.Context)
	GetByUUID(ctx *gin.Context)
	Create(ctx *gin.Context)
	Webhook(ctx *gin.Context)
}

func NewPaymentController(service services.IServiceRegistry) IPaymentController {
	return &PaymentController{service: service}
}

type PaymentController struct {
	service services.IServiceRegistry
}

func (p *PaymentController) GetAllWithPagination(ctx *gin.Context) {
	var param dto.PaymentRequestParam

	if !bindAndValidateRequest(ctx, &param, ctx.ShouldBindQuery) {
		return
	}

	result, err := p.service.GetPayment().GetAllWithPagination(ctx, &param)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Data: result,
		Gin:  ctx,
		Code: http.StatusOK,
	})
}

func (p *PaymentController) GetByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	result, err := p.service.GetPayment().GetByUUID(ctx, uuid)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Data: result,
		Gin:  ctx,
		Code: http.StatusOK,
	})
}

func (p *PaymentController) Create(ctx *gin.Context) {
	var request dto.PaymentRequest
	if !bindAndValidateRequest(ctx, &request, ctx.ShouldBindJSON) {
		return
	}
	result, err := p.service.GetPayment().Create(ctx, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Data: result,
		Gin:  ctx,
		Code: http.StatusCreated,
	})
}

func (p *PaymentController) Webhook(ctx *gin.Context) {
	var request dto.Webhook
	if !bindRequest(ctx, &request, ctx.ShouldBindJSON) {
		return
	}
	err := p.service.GetPayment().Webhook(ctx, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,

			Gin: ctx,
		})
		return
	}
	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Gin:  ctx,
	})
}
