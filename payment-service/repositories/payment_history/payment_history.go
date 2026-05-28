package repositories

import (
	"context"
	errWrap "payment-service/common/error"
	errConstant "payment-service/constants/error"
	"payment-service/domain/dto"
	"payment-service/domain/models"

	"gorm.io/gorm"
)

type IPaymentHistoryRepository interface {
	Create(context.Context, *dto.PaymentHistoryRequest) error
}

func NewPaymentHistoryRepository(db *gorm.DB) IPaymentHistoryRepository {
	return &PaymentHistoryRepository{db: db}
}

type PaymentHistoryRepository struct {
	db *gorm.DB
}

func (p *PaymentHistoryRepository) Create(ctx context.Context, request *dto.PaymentHistoryRequest) error {
	paymentHistory := models.PaymentHistory{
		PaymentID: request.PaymentId,
		Status:    request.Status,
	}

	err := p.db.WithContext(ctx).Create(&paymentHistory).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}
