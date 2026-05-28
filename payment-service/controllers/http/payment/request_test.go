package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"payment-service/common/response"
)

type testRequest struct {
	Name string `validate:"required"`
}

func TestBindAndValidateRequestReturnsFalseAndBadRequestOnBindError(t *testing.T) {
	ctx, recorder := newTestContext()

	ok := bindAndValidateRequest(ctx, &testRequest{}, func(interface{}) error {
		return errors.New("invalid request")
	})

	if ok {
		t.Fatalf("expected binding to fail")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	var body response.Response
	decodeResponse(t, recorder, &body)
	if body.Status != "error" {
		t.Fatalf("expected error status, got %s", body.Status)
	}
}

func TestBindAndValidateRequestReturnsFalseAndValidationResponse(t *testing.T) {
	ctx, recorder := newTestContext()

	ok := bindAndValidateRequest(ctx, &testRequest{}, func(interface{}) error {
		return nil
	})

	if ok {
		t.Fatalf("expected validation to fail")
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, recorder.Code)
	}
	var body response.Response
	decodeResponse(t, recorder, &body)
	if body.Status != "error" {
		t.Fatalf("expected error status, got %s", body.Status)
	}
	if body.Message != http.StatusText(http.StatusUnprocessableEntity) {
		t.Fatalf("expected message %q, got %q", http.StatusText(http.StatusUnprocessableEntity), body.Message)
	}
	data, ok := body.Data.([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("expected one validation error, got %#v", body.Data)
	}
}

func TestBindAndValidateRequestReturnsTrueOnValidRequest(t *testing.T) {
	ctx, _ := newTestContext()
	request := testRequest{Name: "valid"}

	ok := bindAndValidateRequest(ctx, &request, func(interface{}) error {
		return nil
	})

	if !ok {
		t.Fatalf("expected valid request to pass")
	}
}

func TestBindRequestSkipsValidation(t *testing.T) {
	ctx, _ := newTestContext()

	ok := bindRequest(ctx, &testRequest{}, func(interface{}) error {
		return nil
	})

	if !ok {
		t.Fatalf("expected bind request to pass without validation")
	}
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	return ctx, recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}
