package request

import (
	errWrap "field-service/common/error"
	"field-service/common/response"
	errConstant "field-service/constants/error"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func BindQuery(c *gin.Context, target any) bool {
	return bindAndValidate(c, target, c.ShouldBindQuery)
}

func BindJSON(c *gin.Context, target any) bool {
	return bindAndValidate(c, target, c.ShouldBindJSON)
}

func BindWith(c *gin.Context, target any, binding binding.Binding) bool {
	return bindAndValidate(c, target, func(target any) error {
		return c.ShouldBindWith(target, binding)
	})
}

func BindMultipart(c *gin.Context, target any) bool {
	return BindWith(c, target, binding.FormMultipart)
}

func HandleError(c *gin.Context, err error) {
	response.HttpResponse(response.ParamHTTPResp{
		Code: errConstant.ErrStatusCode(err),
		Err:  err,
		Gin:  c,
	})
}

func bindAndValidate(c *gin.Context, target any, bind func(any) error) bool {
	if err := bind(target); err != nil {
		HandleError(c, err)
		return false
	}

	// Gin's ShouldBind* respects binding tags only; validate tags require this explicit pass.
	if err := Validate.Struct(target); err != nil {
		message := http.StatusText(http.StatusUnprocessableEntity)
		response.HttpResponse(response.ParamHTTPResp{
			Code:    errConstant.ErrStatusCode(err),
			Err:     err,
			Gin:     c,
			Message: &message,
			Data:    errWrap.ErrValidationResponse(err),
		})
		return false
	}

	return true
}
