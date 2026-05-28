package controllers

import (
	"net/http"
	errValidation "payment-service/common/error"
	"payment-service/common/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type bindFunc func(any) error

var paymentValidator = validator.New()

func bindRequest(ctx *gin.Context, target any, bind bindFunc) bool {
	if err := bind(target); err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return false
	}
	return true
}

func validateRequest(ctx *gin.Context, target any) bool {
	if err := paymentValidator.Struct(target); err != nil {
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

func bindAndValidateRequest(ctx *gin.Context, target any, bind bindFunc) bool {
	return bindRequest(ctx, target, bind) && validateRequest(ctx, target)
}
