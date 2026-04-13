package response

import (
	"field-service/constants"
	errConstant "field-service/constants/error"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Token     *string     `json:"token,omitempty"`
	RequestId string      `json:"requestId,omitempty"`
}

type ParamHTTPResp struct {
	Code    int
	Err     error
	Message *string
	Gin     *gin.Context
	Data    interface{}
	Token   *string
}

func HttpResponse(param ParamHTTPResp) {
	requestID, _ := param.Gin.Get("request_id")
	requestIDStr := fmt.Sprintf("%v", requestID)

	if param.Err == nil {
		param.Gin.JSON(param.Code, Response{
			Status:    constants.Success,
			Message:   http.StatusText(http.StatusOK),
			Data:      param.Data,
			Token:     param.Token,
			RequestId: requestIDStr,
		})
		return
	}

	message := errConstant.ErrInternalServerError.Error()

	if param.Message != nil {
		message = *param.Message
	} else if param.Err != nil {
		if errConstant.ErrMapping(param.Err) {
			message = param.Err.Error()
		}
	}

	param.Gin.JSON(param.Code, Response{
		Status:    constants.Error,
		Message:   message,
		Data:      param.Data,
		RequestId: requestIDStr,
	})
	return
}

func WithRequestId(resp Response, requestID string) Response {
	resp.RequestId = requestID
	return resp
}
