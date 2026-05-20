package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	errValidation "order-service/common/error"
	"order-service/common/response"
	errConstant "order-service/constants/error"
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

// GetAllWithPagination godoc
// @Summary      Get all orders with pagination
// @Description  Retrieve a paginated list of all orders with optional sorting
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        page        query     int     true   "Page number"
// @Param        limit       query     int     true   "Page limit"
// @Param        sortColumn  query     string  false  "Sort column"
// @Param        sortOrder   query     string  false  "Sort order"
// @Success      200         {object}  response.Response
// @Failure      400         {object}  response.Response
// @Failure      401         {object}  response.Response
// @Failure      500         {object}  response.Response
// @Router       /order [get]
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
				Code: errConstant.ErrStatusCode(err),
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

// GetByUUID godoc
// @Summary      Get order by UUID
// @Description  Retrieve order details by UUID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        uuid        path      string  true   "Order UUID"
// @Success      200         {object}  response.Response
// @Failure      400         {object}  response.Response
// @Failure      401         {object}  response.Response
// @Failure      404         {object}  response.Response
// @Failure      500         {object}  response.Response
// @Router       /order/{uuid} [get]
func (o *OrderController) GetByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	result, err := o.service.GetOrder().GetByUUID(ctx, uuid)
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: errConstant.ErrStatusCode(err),
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

// GetOrderByUserId godoc
// @Summary      Get orders by user ID
// @Description  Retrieve a list of orders for the currently authenticated user
// @Tags         orders
// @Accept       json
// @Produce      json
// @Success      200         {object}  response.Response
// @Failure      400         {object}  response.Response
// @Failure      401         {object}  response.Response
// @Failure      500         {object}  response.Response
// @Router       /order/user [get]
func (o *OrderController) GetOrderByUserId(ctx *gin.Context) {
	result, err := o.service.GetOrder().GetOrderByUserID(ctx.Request.Context())
	if err != nil {
		response.HttpResponse(
			response.ParamHTTPResp{
				Code: errConstant.ErrStatusCode(err),
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

// Create godoc
// @Summary      Create a new order
// @Description  Create a new booking order for soccer field schedules
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request     body      dto.OrderRequest  true  "Order details"
// @Success      200         {object}  response.Response
// @Failure      400         {object}  response.Response
// @Failure      401         {object}  response.Response
// @Failure      422         {object}  response.Response
// @Failure      500         {object}  response.Response
// @Router       /order [post]
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
				Code: errConstant.ErrStatusCode(err),
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
