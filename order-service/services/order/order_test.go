package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	clientField "order-service/clients/field"
	clientPayment "order-service/clients/payment"
	clientUser "order-service/clients/user"
	"order-service/constants"
	errConst "order-service/constants/error"
	errOrder "order-service/constants/error/order"
	"order-service/domain/dto"
	"order-service/domain/models"
)

func setupTest(t *testing.T) (
	*OrderService,
	*MockRepositoryRegistry,
	*MockClientRegistry,
	*MockOrderRepository,
	*MockOrderHistoryRepository,
	*MockOrderFieldRepository,
	*MockUserClient,
	*MockPaymentClient,
	*MockFieldClient,
	sqlmock.Sqlmock,
) {
	db, mockSql, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	mockRepoReg := new(MockRepositoryRegistry)
	mockClientReg := new(MockClientRegistry)

	mockOrderRepo := new(MockOrderRepository)
	mockHistoryRepo := new(MockOrderHistoryRepository)
	mockFieldRepo := new(MockOrderFieldRepository)

	mockUserClient := new(MockUserClient)
	mockPaymentClient := new(MockPaymentClient)
	mockFieldClient := new(MockFieldClient)

	mockRepoReg.On("GetOrder").Return(mockOrderRepo)
	mockRepoReg.On("GetOrderHistory").Return(mockHistoryRepo)
	mockRepoReg.On("GetOrderField").Return(mockFieldRepo)
	mockRepoReg.On("GetTx").Return(gormDB)

	mockClientReg.On("GetUser").Return(mockUserClient)
	mockClientReg.On("GetPayment").Return(mockPaymentClient)
	mockClientReg.On("GetField").Return(mockFieldClient)

	svc := &OrderService{
		repository: mockRepoReg,
		client:     mockClientReg,
	}

	return svc, mockRepoReg, mockClientReg, mockOrderRepo, mockHistoryRepo, mockFieldRepo, mockUserClient, mockPaymentClient, mockFieldClient, mockSql
}

func TestGetAllWithPagination_Success(t *testing.T) {
	svc, _, _, mockOrderRepo, _, _, mockUserClient, _, _, _ := setupTest(t)

	ctx := context.Background()
	param := &dto.OrderRequestParam{
		Page:  1,
		Limit: 10,
	}

	userUUID := uuid.New()
	orderTime := time.Now()
	orders := []models.Order{
		{
			ID:        1,
			UUID:      uuid.New(),
			Code:      "ORD-00001-20260520",
			UserID:    userUUID,
			Amount:    150000,
			Status:    constants.Pending,
			Date:      orderTime,
			CreatedAt: &orderTime,
			UpdatedAt: &orderTime,
		},
	}

	mockOrderRepo.On("FindAllWithPagination", ctx, param).Return(orders, int64(1), nil)
	mockUserClient.On("GetUserByUUID", ctx, userUUID).Return(&clientUser.UserData{
		Name: "Test User",
	}, nil)

	res, err := svc.GetAllWithPagination(ctx, param)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(1), res.TotalData)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 10, res.Limit)
	mockOrderRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestGetAllWithPagination_DBError(t *testing.T) {
	svc, _, _, mockOrderRepo, _, _, _, _, _, _ := setupTest(t)

	ctx := context.Background()
	param := &dto.OrderRequestParam{
		Page:  1,
		Limit: 10,
	}

	dbErr := errors.New("db error")
	mockOrderRepo.On("FindAllWithPagination", ctx, param).Return(nil, int64(0), dbErr)

	res, err := svc.GetAllWithPagination(ctx, param)
	assert.ErrorIs(t, err, dbErr)
	assert.Nil(t, res)
	mockOrderRepo.AssertExpectations(t)
}

func TestGetAllWithPagination_UserClientError(t *testing.T) {
	svc, _, _, mockOrderRepo, _, _, mockUserClient, _, _, _ := setupTest(t)

	ctx := context.Background()
	param := &dto.OrderRequestParam{
		Page:  1,
		Limit: 10,
	}

	userUUID := uuid.New()
	orderTime := time.Now()
	orders := []models.Order{
		{
			ID:        1,
			UUID:      uuid.New(),
			Code:      "ORD-00001-20260520",
			UserID:    userUUID,
			Amount:    150000,
			Status:    constants.Pending,
			Date:      orderTime,
			CreatedAt: &orderTime,
			UpdatedAt: &orderTime,
		},
	}

	clientErr := errors.New("client error")
	mockOrderRepo.On("FindAllWithPagination", ctx, param).Return(orders, int64(1), nil)
	mockUserClient.On("GetUserByUUID", ctx, userUUID).Return(nil, clientErr)

	res, err := svc.GetAllWithPagination(ctx, param)
	assert.ErrorIs(t, err, clientErr)
	assert.Nil(t, res)
	mockOrderRepo.AssertExpectations(t)
	mockUserClient.AssertExpectations(t)
}

func TestCreate_Success(t *testing.T) {
	svc, _, _, mockOrderRepo, mockHistoryRepo, mockFieldRepo, _, mockPaymentClient, mockFieldClient, mockSql := setupTest(t)

	userUUID := uuid.New()
	userData := &clientUser.UserData{
		UUID:        userUUID,
		Name:        "Test User",
		Email:       "test@example.com",
		PhoneNumber: "08123456789",
	}

	ctx := context.WithValue(context.Background(), constants.User, userData)

	fieldScheduleID := uuid.New()
	fieldData := &clientField.FieldData{
		UUID:          fieldScheduleID,
		PricePerHours: 100000,
		Status:        "available",
		FieldName:     "Gelora Bung Karno",
	}

	mockFieldClient.On("GetFieldByUUID", ctx, fieldScheduleID).Return(fieldData, nil)

	createdOrderTime := time.Now()
	createdOrder := &models.Order{
		ID:        1,
		UUID:      uuid.New(),
		Code:      "ORD-00001-20260520",
		UserID:    userUUID,
		Amount:    100000,
		Status:    constants.Pending,
		Date:      createdOrderTime,
		CreatedAt: &createdOrderTime,
		UpdatedAt: &createdOrderTime,
	}

	mockSql.ExpectBegin()
	mockOrderRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(createdOrder, nil)
	mockFieldRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	mockHistoryRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)

	paymentUUID := uuid.New()
	mockPaymentClient.On("CreatePaymentLink", ctx, mock.Anything).Return(&clientPayment.PaymentData{
		UUID:        paymentUUID,
		PaymentLink: "https://midtrans.com/pay/123",
	}, nil)

	mockOrderRepo.On("Update", ctx, mock.Anything, createdOrder.UUID, mock.Anything).Return(nil)
	mockSql.ExpectCommit()

	param := &dto.OrderRequest{
		FieldScheduleIDs: []string{fieldScheduleID.String()},
	}

	res, err := svc.Create(ctx, param)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, createdOrder.UUID, res.UUID)
	assert.Equal(t, "https://midtrans.com/pay/123", res.PaymentLink)
	assert.NoError(t, mockSql.ExpectationsWereMet())
	mockOrderRepo.AssertExpectations(t)
	mockFieldClient.AssertExpectations(t)
	mockPaymentClient.AssertExpectations(t)
}

func TestCreate_Unauthorized(t *testing.T) {
	svc, _, _, _, _, _, _, _, _, _ := setupTest(t)

	ctx := context.Background() // no user context
	param := &dto.OrderRequest{
		FieldScheduleIDs: []string{uuid.New().String()},
	}

	res, err := svc.Create(ctx, param)
	assert.ErrorIs(t, err, errConst.ErrUnauthorized)
	assert.Nil(t, res)
}

func TestCreate_FieldAlreadyBooked(t *testing.T) {
	svc, _, _, _, _, _, _, _, mockFieldClient, _ := setupTest(t)

	userUUID := uuid.New()
	userData := &clientUser.UserData{
		UUID: userUUID,
	}
	ctx := context.WithValue(context.Background(), constants.User, userData)

	fieldScheduleID := uuid.New()
	fieldData := &clientField.FieldData{
		UUID:          fieldScheduleID,
		PricePerHours: 100000,
		Status:        "booked", // Already booked!
		FieldName:     "Gelora Bung Karno",
	}

	mockFieldClient.On("GetFieldByUUID", ctx, fieldScheduleID).Return(fieldData, nil)

	param := &dto.OrderRequest{
		FieldScheduleIDs: []string{fieldScheduleID.String()},
	}

	res, err := svc.Create(ctx, param)
	assert.ErrorIs(t, err, errOrder.ErrFieldAlreadyBooked)
	assert.Nil(t, res)
}

func TestCreate_PaymentClientError(t *testing.T) {
	svc, _, _, mockOrderRepo, mockHistoryRepo, mockFieldRepo, _, mockPaymentClient, mockFieldClient, mockSql := setupTest(t)

	userUUID := uuid.New()
	userData := &clientUser.UserData{
		UUID:        userUUID,
		Name:        "Test User",
		Email:       "test@example.com",
		PhoneNumber: "08123456789",
	}
	ctx := context.WithValue(context.Background(), constants.User, userData)

	fieldScheduleID := uuid.New()
	fieldData := &clientField.FieldData{
		UUID:          fieldScheduleID,
		PricePerHours: 100000,
		Status:        "available",
		FieldName:     "Gelora Bung Karno",
	}

	mockFieldClient.On("GetFieldByUUID", ctx, fieldScheduleID).Return(fieldData, nil)

	createdOrderTime := time.Now()
	createdOrder := &models.Order{
		ID:        1,
		UUID:      uuid.New(),
		Code:      "ORD-00001-20260520",
		UserID:    userUUID,
		Amount:    100000,
		Status:    constants.Pending,
		Date:      createdOrderTime,
		CreatedAt: &createdOrderTime,
		UpdatedAt: &createdOrderTime,
	}

	mockSql.ExpectBegin()
	mockOrderRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(createdOrder, nil)
	mockFieldRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	mockHistoryRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)

	paymentErr := errors.New("payment service down")
	mockPaymentClient.On("CreatePaymentLink", ctx, mock.Anything).Return(nil, paymentErr)
	mockSql.ExpectRollback()

	param := &dto.OrderRequest{
		FieldScheduleIDs: []string{fieldScheduleID.String()},
	}

	res, err := svc.Create(ctx, param)
	assert.ErrorIs(t, err, paymentErr)
	assert.Nil(t, res)
	assert.NoError(t, mockSql.ExpectationsWereMet())
	mockOrderRepo.AssertExpectations(t)
	mockFieldClient.AssertExpectations(t)
	mockPaymentClient.AssertExpectations(t)
}

func TestHandlePayment_Settlement(t *testing.T) {
	svc, _, _, mockOrderRepo, mockHistoryRepo, mockFieldRepo, _, _, mockFieldClient, mockSql := setupTest(t)

	ctx := context.Background()
	orderUUID := uuid.New()
	paymentUUID := uuid.New()
	paidTime := time.Now()

	paymentData := &dto.PaymentData{
		OrderID:   orderUUID,
		PaymentID: paymentUUID,
		Status:    constants.SettlementPaymentStatus,
		PaidAt:    &paidTime,
	}

	order := &models.Order{
		ID:        1,
		UUID:      orderUUID,
		Code:      "ORD-00001-20260520",
		UserID:    uuid.New(),
		Amount:    100000,
		Status:    constants.Pending,
		Date:      time.Now(),
	}

	orderFieldSchedules := []models.OrderField{
		{
			ID:              1,
			OrderID:         1,
			FieldScheduleID: uuid.New(),
		},
	}

	mockSql.ExpectBegin()
	mockOrderRepo.On("Update", ctx, mock.Anything, orderUUID, mock.Anything).Return(nil)
	mockOrderRepo.On("FindByUUID", ctx, orderUUID.String()).Return(order, nil)
	mockHistoryRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	mockFieldRepo.On("FindByOrderID", ctx, uint(1)).Return(orderFieldSchedules, nil)
	mockFieldClient.On("UpdateStatus", mock.Anything).Return(nil)
	mockSql.ExpectCommit()

	err := svc.HandlePayment(ctx, paymentData)
	assert.NoError(t, err)
	assert.NoError(t, mockSql.ExpectationsWereMet())
	mockOrderRepo.AssertExpectations(t)
	mockHistoryRepo.AssertExpectations(t)
	mockFieldRepo.AssertExpectations(t)
	mockFieldClient.AssertExpectations(t)
}

func TestHandlePayment_Expired(t *testing.T) {
	svc, _, _, mockOrderRepo, mockHistoryRepo, _, _, _, _, mockSql := setupTest(t)

	ctx := context.Background()
	orderUUID := uuid.New()
	paymentUUID := uuid.New()

	paymentData := &dto.PaymentData{
		OrderID:   orderUUID,
		PaymentID: paymentUUID,
		Status:    constants.ExpirePaymentStatus,
	}

	order := &models.Order{
		ID:        1,
		UUID:      orderUUID,
		Code:      "ORD-00001-20260520",
		UserID:    uuid.New(),
		Amount:    100000,
		Status:    constants.Pending,
		Date:      time.Now(),
	}

	mockSql.ExpectBegin()
	mockOrderRepo.On("Update", ctx, mock.Anything, orderUUID, mock.Anything).Return(nil)
	mockOrderRepo.On("FindByUUID", ctx, orderUUID.String()).Return(order, nil)
	mockHistoryRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	mockSql.ExpectCommit()

	err := svc.HandlePayment(ctx, paymentData)
	assert.NoError(t, err)
	assert.NoError(t, mockSql.ExpectationsWereMet())
	mockOrderRepo.AssertExpectations(t)
	mockHistoryRepo.AssertExpectations(t)
}

func TestHandlePayment_UnknownStatus(t *testing.T) {
	svc, _, _, _, _, _, _, _, _, _ := setupTest(t)

	ctx := context.Background()
	paymentData := &dto.PaymentData{
		OrderID:   uuid.New(),
		PaymentID: uuid.New(),
		Status:    "deny", // unknown / unsupported status
	}

	err := svc.HandlePayment(ctx, paymentData)
	assert.ErrorIs(t, err, errOrder.ErrUnknownPaymentStatus)
}
