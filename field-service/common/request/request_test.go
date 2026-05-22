package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errConstant "field-service/constants/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testRequest struct {
	Name string `form:"name" json:"name" validate:"required"`
}

func TestBindJSONReturnsValidationResponse(t *testing.T) {
	c, w := newTestContext()
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var request testRequest
	ok := BindJSON(c, &request)

	assert.False(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "error", body["status"])
	assert.Equal(t, http.StatusText(http.StatusUnprocessableEntity), body["message"])
	assert.NotEmpty(t, body["data"])
}

func TestBindJSONReturnsTrueForValidRequest(t *testing.T) {
	c, w := newTestContext()
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"field"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	var request testRequest
	ok := BindJSON(c, &request)

	assert.True(t, ok)
	assert.Equal(t, "field", request.Name)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBindJSONReturnsBindErrorForMalformedInput(t *testing.T) {
	c, w := newTestContext()
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":`))
	c.Request.Header.Set("Content-Type", "application/json")

	var request testRequest
	ok := BindJSON(c, &request)

	assert.False(t, ok)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBindQueryReturnsTrueForValidRequest(t *testing.T) {
	c, w := newTestContext()
	c.Request = httptest.NewRequest(http.MethodGet, "/?name=field", nil)

	var request testRequest
	ok := BindQuery(c, &request)

	assert.True(t, ok)
	assert.Equal(t, "field", request.Name)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBindMultipartReturnsTrueForValidRequest(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err := writer.WriteField("name", "field")
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	c, w := newTestContext()
	c.Request = httptest.NewRequest(http.MethodPost, "/", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	var request testRequest
	ok := BindMultipart(c, &request)

	assert.True(t, ok)
	assert.Equal(t, "field", request.Name)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleErrorWritesMappedResponse(t *testing.T) {
	c, w := newTestContext()

	HandleError(c, errConstant.ErrUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleErrorWritesUnknownErrorAsBadRequest(t *testing.T) {
	c, w := newTestContext()

	HandleError(c, errors.New("unknown error"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}
