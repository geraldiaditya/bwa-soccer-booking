package repositories

import (
	"context"
	"gorm.io/gorm"
	errWrap "payment-service/common/error"
	errConstant "payment-service/constants/error"
	"payment-service/domain/dto"
	"payment-service/domain/models"
)

type IPaymentHistoryRepository interface {
	Create(context.Context, *gorm.DB, *dto.PaymentHistoryRequest) error
}

func NewPaymentHistoryRepository(db *gorm.DB) IPaymentHistoryRepository {
	return &PaymentHistoryRepository{db: db}
}

type PaymentHistoryRepository struct {
	db *gorm.DB
}

func (p *PaymentHistoryRepository) Create(ctx context.Context, db *gorm.DB, request *dto.PaymentHistoryRequest) error {
	paymentHistory := models.PaymentHistory{
		PaymentID: request.PaymentId,
		Status:    request.Status,
	}

	err := db.WithContext(ctx).Create(&paymentHistory).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}
	return nil
}
