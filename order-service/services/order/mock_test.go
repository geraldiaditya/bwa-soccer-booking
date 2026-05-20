package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	clientField "order-service/clients/field"
	clientPayment "order-service/clients/payment"
	clientUser "order-service/clients/user"
	"order-service/domain/dto"
	"order-service/domain/models"
	orderRepo "order-service/repositories/order"
	orderFieldRepo "order-service/repositories/order_field"
	orderHistoryRepo "order-service/repositories/order_history"
)

// MockRepositoryRegistry
type MockRepositoryRegistry struct {
	mock.Mock
}

func (m *MockRepositoryRegistry) GetOrder() orderRepo.IOrderRepository {
	args := m.Called()
	return args.Get(0).(orderRepo.IOrderRepository)
}

func (m *MockRepositoryRegistry) GetOrderHistory() orderHistoryRepo.IOrderHistoryRepository {
	args := m.Called()
	return args.Get(0).(orderHistoryRepo.IOrderHistoryRepository)
}

func (m *MockRepositoryRegistry) GetOrderField() orderFieldRepo.IOrderFieldRepository {
	args := m.Called()
	return args.Get(0).(orderFieldRepo.IOrderFieldRepository)
}

func (m *MockRepositoryRegistry) GetTx() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// MockOrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) FindAllWithPagination(ctx context.Context, param *dto.OrderRequestParam) ([]models.Order, int64, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByUUID(ctx context.Context, uuid string) (*models.Order, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) Create(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error) {
	args := m.Called(ctx, tx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, tx *gorm.DB, uuid uuid.UUID, order *models.Order) error {
	args := m.Called(ctx, tx, uuid, order)
	return args.Error(0)
}

// MockOrderFieldRepository
type MockOrderFieldRepository struct {
	mock.Mock
}

func (m *MockOrderFieldRepository) FindByOrderID(ctx context.Context, orderID uint) ([]models.OrderField, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OrderField), args.Error(1)
}

func (m *MockOrderFieldRepository) Create(ctx context.Context, tx *gorm.DB, requests []models.OrderField) error {
	args := m.Called(ctx, tx, requests)
	return args.Error(0)
}

// MockOrderHistoryRepository
type MockOrderHistoryRepository struct {
	mock.Mock
}

func (m *MockOrderHistoryRepository) Create(ctx context.Context, tx *gorm.DB, request *dto.OrderHistoryRequest) error {
	args := m.Called(ctx, tx, request)
	return args.Error(0)
}

// MockClientRegistry
type MockClientRegistry struct {
	mock.Mock
}

func (m *MockClientRegistry) GetUser() clientUser.IUserClient {
	args := m.Called()
	return args.Get(0).(clientUser.IUserClient)
}

func (m *MockClientRegistry) GetPayment() clientPayment.IPaymentClient {
	args := m.Called()
	return args.Get(0).(clientPayment.IPaymentClient)
}

func (m *MockClientRegistry) GetField() clientField.IFieldClient {
	args := m.Called()
	return args.Get(0).(clientField.IFieldClient)
}

// MockUserClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUserByToken(ctx context.Context) (*clientUser.UserData, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clientUser.UserData), args.Error(1)
}

func (m *MockUserClient) GetUserByUUID(ctx context.Context, uuid uuid.UUID) (*clientUser.UserData, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clientUser.UserData), args.Error(1)
}

// MockPaymentClient
type MockPaymentClient struct {
	mock.Mock
}

func (m *MockPaymentClient) GetPaymentByUUID(ctx context.Context, uuid uuid.UUID) (*clientPayment.PaymentData, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clientPayment.PaymentData), args.Error(1)
}

func (m *MockPaymentClient) CreatePaymentLink(ctx context.Context, request *dto.PaymentRequest) (*clientPayment.PaymentData, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clientPayment.PaymentData), args.Error(1)
}

// MockFieldClient
type MockFieldClient struct {
	mock.Mock
}

func (m *MockFieldClient) GetFieldByUUID(ctx context.Context, uuid uuid.UUID) (*clientField.FieldData, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clientField.FieldData), args.Error(1)
}

func (m *MockFieldClient) UpdateStatus(request *dto.UpdateStatusFieldScheduleRequest) error {
	args := m.Called(request)
	return args.Error(0)
}
