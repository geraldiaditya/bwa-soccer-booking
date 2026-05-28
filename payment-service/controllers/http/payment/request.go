package controllers

import (
	"net/http"
	errValidation "payment-service/common/error"
	"payment-service/common/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type requestBinder func(interface{}) error

func bindRequest(ctx *gin.Context, request interface{}, binder requestBinder) bool {
	if err := binder(request); err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return false
	}
	return true
}

func bindAndValidateRequest(ctx *gin.Context, request interface{}, binder requestBinder) bool {
	if !bindRequest(ctx, request, binder) {
		return false
	}
	if err := validator.New().Struct(request); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errResponse := errValidation.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Err:     err,
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errResponse,
			Gin:     ctx,
		})
		return false
	}
	return true
}
