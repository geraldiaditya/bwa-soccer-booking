package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"payment-service/common/response"
	"payment-service/domain/dto"
)

func TestBindRequestWritesBadRequestOnBindError(t *testing.T) {
	ctx, recorder := newTestContext("POST", "/payment", strings.NewReader("{"))
	ctx.Request.Header.Set("Content-Type", "application/json")
	var request dto.PaymentRequest

	if bindRequest(ctx, &request, ctx.ShouldBindJSON) {
		t.Fatal("expected bindRequest to fail")
	}

	assertStatus(t, http.StatusBadRequest, recorder.Code)
	assertErrorResponse(t, recorder.Body.String(), "internal server error")
}

func TestBindAndValidateRequestWritesValidationError(t *testing.T) {
	ctx, recorder := newTestContext("GET", "/payment?limit=10", nil)
	var param dto.PaymentRequestParam

	if bindAndValidateRequest(ctx, &param, ctx.ShouldBindQuery) {
		t.Fatal("expected bindAndValidateRequest to fail")
	}

	assertStatus(t, http.StatusUnprocessableEntity, recorder.Code)
	assertErrorResponse(t, recorder.Body.String(), http.StatusText(http.StatusUnprocessableEntity))
}

func TestBindAndValidateRequestReturnsTrueForValidRequest(t *testing.T) {
	ctx, recorder := newTestContext("GET", "/payment?page=1&limit=10", nil)
	var param dto.PaymentRequestParam

	if !bindAndValidateRequest(ctx, &param, ctx.ShouldBindQuery) {
		t.Fatal("expected bindAndValidateRequest to succeed")
	}

	assertStatus(t, http.StatusOK, recorder.Code)
	if param.Page != 1 || param.Limit != 10 {
		t.Fatalf("expected bound page and limit, got page=%d limit=%d", param.Page, param.Limit)
	}
}

func TestValidateRequestWritesValidationResponse(t *testing.T) {
	ctx, recorder := newTestContext("GET", "/payment", nil)
	param := dto.PaymentRequestParam{Limit: 10}

	if validateRequest(ctx, &param) {
		t.Fatal("expected validateRequest to fail")
	}

	assertStatus(t, http.StatusUnprocessableEntity, recorder.Code)
	var body response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	items, ok := body.Data.([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected validation data, got %#v", body.Data)
	}
}

func TestBindRequestReturnsTrueForSuccessfulBinder(t *testing.T) {
	ctx, recorder := newTestContext("POST", "/payment", nil)

	if !bindRequest(ctx, &struct{}{}, func(any) error { return nil }) {
		t.Fatal("expected bindRequest to succeed")
	}

	assertStatus(t, http.StatusOK, recorder.Code)
}

func TestBindRequestReturnsFalseForCustomBinderError(t *testing.T) {
	ctx, recorder := newTestContext("POST", "/payment", nil)

	if bindRequest(ctx, &struct{}{}, func(any) error { return errors.New("bind failed") }) {
		t.Fatal("expected bindRequest to fail")
	}

	assertStatus(t, http.StatusBadRequest, recorder.Code)
}

func newTestContext(method, target string, body *strings.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	var requestBody *strings.Reader
	if body == nil {
		requestBody = strings.NewReader("")
	} else {
		requestBody = body
	}
	ctx.Request = httptest.NewRequest(method, target, requestBody)
	return ctx, recorder
}

func assertStatus(t *testing.T, expected, actual int) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected status %d, got %d", expected, actual)
	}
}

func assertErrorResponse(t *testing.T, body, expectedMessage string) {
	t.Helper()
	var responseBody response.Response
	if err := json.Unmarshal([]byte(body), &responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if responseBody.Status != "error" {
		t.Fatalf("expected error status, got %q", responseBody.Status)
	}
	if responseBody.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, responseBody.Message)
	}
}
