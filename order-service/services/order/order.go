package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"order-service/clients"
	clientField "order-service/clients/field"
	clientPayment "order-service/clients/payment"
	clientUser "order-service/clients/user"
	"order-service/common/utils"
	"order-service/constants"
	errOrder "order-service/constants/error/order"
	"order-service/domain/dto"
	"order-service/domain/models"
	repo "order-service/repositories"
	"time"
)

func NewOrderService(repository repo.IRepositoryRegistry, client clients.IClientRegistry) IOrderService {
	return &OrderService{
		repository: repository,
		client:     client,
	}
}

type IOrderService interface {
	GetAllWithPagination(ctx context.Context, param *dto.OrderRequestParam) (*utils.PaginationResult, error)
	GetByUUID(ctx context.Context, id string) (*dto.OrderResponse, error)
	GetOrderByUserID(ctx context.Context) ([]dto.OrderByUserIDResponse, error)
	Create(ctx context.Context, param *dto.OrderRequest) (*dto.OrderResponse, error)
	HandlePayment(ctx context.Context, data *dto.PaymentData) error
}

type OrderService struct {
	repository repo.IRepositoryRegistry
	client     clients.IClientRegistry
}

func (o *OrderService) GetAllWithPagination(ctx context.Context, param *dto.OrderRequestParam) (*utils.PaginationResult, error) {
	orders, total, err := o.repository.GetOrder().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}
	orderResult := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		resUser, err := o.client.GetUser().GetUserByUUID(ctx, order.UserID)
		if err != nil {
			return nil, err
		}
		orderResult = append(orderResult, dto.OrderResponse{
			UUID:      order.UUID,
			Code:      order.Code,
			UserName:  resUser.Name,
			Amount:    order.Amount,
			Status:    order.Status.GetStatusString(),
			OrderDate: order.Date,
			CreatedAt: *order.CreatedAt,
			UpdateAt:  *order.UpdatedAt,
		})
	}

	paginationParam := utils.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  orderResult,
	}
	response := utils.GeneratePagination(paginationParam)
	return &response, nil
}

func (o *OrderService) GetByUUID(ctx context.Context, uuid string) (*dto.OrderResponse, error) {
	var (
		order *models.Order
		user  *clientUser.UserData
		err   error
	)
	order, err = o.repository.GetOrder().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	user, err = o.client.GetUser().GetUserByUUID(ctx, order.UserID)
	if err != nil {
		return nil, err
	}
	response := dto.OrderResponse{
		UUID:      order.UUID,
		Code:      order.Code,
		UserName:  user.Name,
		Amount:    order.Amount,
		Status:    order.Status.GetStatusString(),
		OrderDate: order.Date,
		CreatedAt: *order.CreatedAt,
		UpdateAt:  *order.UpdatedAt,
	}
	return &response, nil
}

func (o *OrderService) GetOrderByUserID(ctx context.Context) ([]dto.OrderByUserIDResponse, error) {
	var (
		orders []models.Order
		err    error
		user   = ctx.Value(constants.User).(*clientUser.UserData)
	)
	orders, err = o.repository.GetOrder().FindByUserID(ctx, user.UUID.String())
	if err != nil {
		return nil, err
	}
	orderLists := make([]dto.OrderByUserIDResponse, 0, len(orders))
	for _, item := range orders {
		payment, err := o.client.GetPayment().GetPaymentByUUID(ctx, item.PaymentID)
		if err != nil {
			return nil, err
		}
		orderLists = append(orderLists, dto.OrderByUserIDResponse{
			Code:        item.Code,
			Amount:      utils.RupiahFormat(&item.Amount),
			Status:      item.Status.GetStatusString(),
			OrderDate:   item.Date.String(),
			PaymentLink: payment.PaymentLink,
			InvoiceLink: payment.InvoiceLink,
		})
	}
	return orderLists, nil
}

func (o *OrderService) Create(ctx context.Context, param *dto.OrderRequest) (*dto.OrderResponse, error) {
	var (
		order               *models.Order
		txErr, err          error
		user                = ctx.Value(constants.User).(*clientUser.UserData)
		field               *clientField.FieldData
		paymentResponse     *clientPayment.PaymentData
		orderFieldSchedules = make([]models.OrderField, 0, len(param.FieldScheduleIDs))
		totalAmount         float64
	)

	for _, fieldID := range param.FieldScheduleIDs {
		uuidParsed := uuid.MustParse(fieldID)
		field, err := o.client.GetField().GetFieldByUUID(ctx, uuidParsed)
		if err != nil {
			return nil, err
		}

		totalAmount += field.PricePerHours
		if field.Status == constants.BookedStatus.String() {
			return nil, errOrder.ErrFieldAlreadyBooked
		}
	}
	err = o.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		order, txErr = o.repository.GetOrder().Create(ctx, tx, &models.Order{
			UserID: user.UUID,
			Amount: totalAmount,
			Date:   time.Now(),
			Status: constants.Pending,
			IsPaid: false,
		})
		if txErr != nil {
			return txErr
		}
		for _, fieldID := range param.FieldScheduleIDs {
			uuidParsed := uuid.MustParse(fieldID)
			orderFieldSchedules = append(orderFieldSchedules, models.OrderField{
				OrderID:         order.ID,
				FieldScheduleID: uuidParsed,
			})
		}
		txErr = o.repository.GetOrderField().Create(ctx, tx, orderFieldSchedules)
		if txErr != nil {
			return txErr
		}
		txErr = o.repository.GetOrderHistory().Create(ctx, tx, &dto.OrderHistoryRequest{
			OrderID: order.ID,
			Status:  constants.Pending.GetStatusString(),
		})
		if txErr != nil {
			return txErr
		}
		expiredAt := time.Now().Add(1 * time.Hour)
		description := fmt.Sprintf("Pembayaran Sewa %s", field.FieldName)
		paymentResponse, txErr = o.client.GetPayment().CreatePaymentLink(ctx, &dto.PaymentRequest{
			PaymentLink: "",
			OrderID:     order.UUID.String(),
			ExpiredAt:   expiredAt,
			Amount:      totalAmount,
			Description: description,
			CustomerDetail: dto.CustomerDetail{
				Name:  user.Name,
				Email: user.Email,
				Phone: user.PhoneNumber,
			},
			ItemDetails: []dto.ItemDetail{
				{
					ID:       uuid.New(),
					Amount:   totalAmount,
					Name:     description,
					Quantity: 1,
				},
			},
		})
		if txErr != nil {
			return txErr
		}
		txErr = o.repository.GetOrder().Update(ctx, tx, order.UUID, &models.Order{
			PaymentID: paymentResponse.UUID,
		})
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	response := dto.OrderResponse{
		UUID:        order.UUID,
		Code:        order.Code,
		Amount:      order.Amount,
		Status:      order.Status.GetStatusString(),
		OrderDate:   order.Date,
		UserName:    user.Name,
		PaymentLink: paymentResponse.PaymentLink,
		CreatedAt:   *order.CreatedAt,
		UpdateAt:    *order.UpdatedAt,
	}
	return &response, nil
}

func (o *OrderService) HandlePayment(ctx context.Context, data *dto.PaymentData) error {
	var (
		err, txErr          error
		order               *models.Order
		orderFieldSchedules []models.OrderField
	)
	status, body := o.mapPaymentStatusToOrder(data)
	err = o.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		txErr = o.repository.GetOrder().Update(ctx, tx, data.OrderID, body)
		if txErr != nil {
			return txErr
		}
		order, txErr = o.repository.GetOrder().FindByUUID(ctx, data.OrderID.String())
		if txErr != nil {
			return txErr
		}
		txErr = o.repository.GetOrderHistory().Create(ctx, tx, &dto.OrderHistoryRequest{
			Status:  status.GetStatusString(),
			OrderID: order.ID,
		})
		if data.Status == constants.SettlementPaymentStatus {
			orderFieldSchedules, txErr = o.repository.GetOrderField().FindByOrderID(ctx, order.ID)
			if txErr != nil {
				return txErr
			}
			fieldSchedulesIDs := make([]string, 0, len(orderFieldSchedules))
			for _, item := range orderFieldSchedules {
				fieldSchedulesIDs = append(fieldSchedulesIDs, item.FieldScheduleID.String())
			}
			txErr = o.client.GetField().UpdateStatus(&dto.UpdateStatusFieldScheduleRequest{
				FieldScheduleIDs: fieldSchedulesIDs,
			})
			if txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *OrderService) mapPaymentStatusToOrder(request *dto.PaymentData) (constants.OrderStatus, *models.Order) {
	var (
		status constants.OrderStatus
		order  *models.Order
	)
	switch request.Status {
	case constants.SettlementPaymentStatus:
		status = constants.PaymentSuccess
		order = &models.Order{
			IsPaid:    true,
			PaymentID: request.PaymentID,
			PaidAt:    request.PaidAt,
			Status:    status,
		}
	case constants.ExpirePaymentStatus:
		status = constants.Expired
		order = &models.Order{
			IsPaid:    false,
			PaymentID: request.PaymentID,
			Status:    status,
		}
	case constants.PendingPaymentStatus:
		status = constants.PendingPayment
		order = &models.Order{
			IsPaid:    false,
			PaymentID: request.PaymentID,
			Status:    status,
		}
	}
	return status, order
}
