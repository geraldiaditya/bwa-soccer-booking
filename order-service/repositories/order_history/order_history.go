package repositories

import (
	"context"
	"gorm.io/gorm"
	errWrap "order-service/common/error"
	errConst "order-service/constants/error"
	"order-service/domain/dto"
	"order-service/domain/models"
)

func NewOrderHistoryRepository(db *gorm.DB) IOrderHistoryRepository {
	return &OrderHistoryRepository{db: db}
}

type OrderHistoryRepository struct {
	db *gorm.DB
}

func (o *OrderHistoryRepository) Create(ctx context.Context, tx *gorm.DB, request *dto.OrderHistoryRequest) error {
	orderHistory := &models.OrderHistory{
		OrderID: request.OrderID,
		Status:  request.Status,
	}
	err := tx.WithContext(ctx).Create(&orderHistory).Error
	if err != nil {
		return errWrap.WrapError(errConst.ErrSQLError)
	}
	return nil
}

type IOrderHistoryRepository interface {
	Create(context.Context, *gorm.DB, *dto.OrderHistoryRequest) error
}
