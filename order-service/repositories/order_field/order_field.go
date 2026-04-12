package repositories

import (
	"context"
	"gorm.io/gorm"
	errWrap "order-service/common/error"
	errConst "order-service/constants/error"
	"order-service/domain/models"
)

func NewOrderFieldRepository(db *gorm.DB) IOrderFieldRepository {
	return &OrderFieldRepository{db: db}
}

type OrderFieldRepository struct {
	db *gorm.DB
}

func (o *OrderFieldRepository) FindByOrderID(ctx context.Context, orderId uint) ([]models.OrderField, error) {
	var orderFields []models.OrderField
	err := o.db.WithContext(ctx).Where("order_id = ?", orderId).Find(&orderFields).Error
	if err != nil {
		return nil, errWrap.WrapError(errConst.ErrSQLError)
	}
	return orderFields, nil
}

func (o *OrderFieldRepository) Create(ctx context.Context, tx *gorm.DB, requests []models.OrderField) error {
	err := tx.WithContext(ctx).Create(&requests).Error
	if err != nil {
		return errWrap.WrapError(errConst.ErrSQLError)
	}
	return nil
}

type IOrderFieldRepository interface {
	FindByOrderID(context.Context, uint) ([]models.OrderField, error)
	Create(context.Context, *gorm.DB, []models.OrderField) error
}
