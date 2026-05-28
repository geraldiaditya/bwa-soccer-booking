package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"payment-service/clients"
	userClient "payment-service/clients/user"
	"payment-service/config"
	paymentController "payment-service/controllers/http/payment"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeControllerRegistry struct{}

func (fakeControllerRegistry) GetPayment() paymentController.IPaymentController {
	return fakePaymentController{}
}

type fakePaymentController struct{}

func (fakePaymentController) GetAllWithPagination(ctx *gin.Context) {}
func (fakePaymentController) GetByUUID(ctx *gin.Context)            {}
func (fakePaymentController) Create(ctx *gin.Context)               {}
func (fakePaymentController) Webhook(ctx *gin.Context)              {}

type fakeClientRegistry struct{}

func (fakeClientRegistry) GetUser() userClient.IUserClient {
	return fakeUserClient{}
}

type fakeUserClient struct{}

func (fakeUserClient) GetUserByToken(context.Context) (*userClient.UserData, error) {
	return &userClient.UserData{}, nil
}

func TestNewRouterRootRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter(config.AppConfig{
		RateLimiterMaxRequest: 10,
		RateLimiterTimeSecond: 1,
	}, dependencies{
		Controller: fakeControllerRegistry{},
		Client:     fakeClientRegistry{},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Status != "success" {
		t.Fatalf("expected success status, got %q", response.Status)
	}
	if response.Message != "Welcome to Payment Service" {
		t.Fatalf("expected welcome message, got %q", response.Message)
	}
}

func TestNewRouterNoRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newRouter(config.AppConfig{
		RateLimiterMaxRequest: 10,
		RateLimiterTimeSecond: 1,
	}, dependencies{
		Controller: fakeControllerRegistry{},
		Client:     fakeClientRegistry{},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

var _ clients.IClientRegistry = fakeClientRegistry{}
