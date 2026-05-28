package repositories

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	errWrap "payment-service/common/error"
	"payment-service/constants"
	errConstant "payment-service/constants/error"
	errPayment "payment-service/constants/error/payment"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"strings"
)

const (
	defaultPaymentLimit = 10
	defaultPaymentSort  = "created_at desc"
)

var allowedPaymentSortColumns = map[string]struct{}{
	"amount":     {},
	"created_at": {},
	"expired_at": {},
	"order_id":   {},
	"paid_at":    {},
	"status":     {},
	"updated_at": {},
}

type IPaymentRepository interface {
	FindAllWithPagination(context.Context, *dto.PaymentRequestParam) ([]models.Payment, int64, error)
	FindByUUID(context.Context, string) (*models.Payment, error)
	FindByOrderID(context.Context, string) (*models.Payment, error)
	Create(context.Context, *dto.PaymentRequest) (*models.Payment, error)
	Update(context.Context, string, *dto.UpdatePaymentRequest) (*models.Payment, error)
}

func NewPaymentRepository(db *gorm.DB) IPaymentRepository {
	return &PaymentRepository{db: db}
}

type PaymentRepository struct {
	db *gorm.DB
}

func (p *PaymentRepository) FindAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) ([]models.Payment, int64, error) {
	var (
		fields []models.Payment
		total  int64
	)

	limit, offset := paymentPagination(param)
	err := p.db.
		WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order(paymentSort(param)).
		Find(&fields).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}
	err = p.db.
		WithContext(ctx).
		Model(&models.Payment{}).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return fields, total, nil
}

func paymentPagination(param *dto.PaymentRequestParam) (int, int) {
	page := param.Page
	if page < 1 {
		page = 1
	}
	limit := param.Limit
	if limit < 1 {
		limit = defaultPaymentLimit
	}
	return limit, (page - 1) * limit
}

func paymentSort(param *dto.PaymentRequestParam) string {
	if param.SortColumn == nil || param.SortOrder == nil {
		return defaultPaymentSort
	}
	column := strings.ToLower(*param.SortColumn)
	if _, ok := allowedPaymentSortColumns[column]; !ok {
		return defaultPaymentSort
	}
	order := "desc"
	if strings.EqualFold(*param.SortOrder, "asc") {
		order = "asc"
	}
	return column + " " + order
}

func (p *PaymentRepository) FindByUUID(ctx context.Context, uuid string) (*models.Payment, error) {
	var payment models.Payment
	err := p.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		First(&payment).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errPayment.ErrPaymentNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &payment, nil
}

func (p *PaymentRepository) FindByOrderID(ctx context.Context, orderId string) (*models.Payment, error) {
	var payment models.Payment
	err := p.db.WithContext(ctx).
		Where("order_id = ?", orderId).
		First(&payment).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errPayment.ErrPaymentNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &payment, nil
}

func (p *PaymentRepository) Create(ctx context.Context, request *dto.PaymentRequest) (*models.Payment, error) {
	status := constants.Initial
	orderId := uuid.MustParse(request.OrderID)
	payment := models.Payment{
		UUID:        uuid.New(),
		OrderID:     orderId,
		Amount:      request.Amount,
		PaymentLink: request.PaymentLink,
		ExpiredAt:   &request.ExpiredAt,
		Description: request.Description,
		Status:      &status,
	}
	err := p.db.WithContext(ctx).Create(&payment).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &payment, nil
}

func (p *PaymentRepository) Update(ctx context.Context, orderId string, request *dto.UpdatePaymentRequest) (*models.Payment, error) {
	payment := models.Payment{
		Status:        request.Status,
		TransactionID: request.TransactionId,
		InvoiceLink:   request.InvoiceLink,
		PaidAt:        request.PaidAt,
		VANumber:      request.VANumber,
		Bank:          request.Bank,
		Acquirer:      &request.Acquirer,
	}
	err := p.db.WithContext(ctx).Where("order_id = ?", orderId).Updates(&payment).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}
	return &payment, nil
}
