package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	errWrap "order-service/common/error"
	errConst "order-service/constants/error"
	errOrder "order-service/constants/error/order"
	"order-service/domain/dto"
	"order-service/domain/models"
	"strconv"
	"time"
)

func NewOrderRepository(db *gorm.DB) IOrderRepository {
	return &OrderRepository{db: db}
}

type OrderRepository struct {
	db *gorm.DB
}

func (o *OrderRepository) FindByUUID(ctx context.Context, uuid string) (*models.Order, error) {
	var order models.Order
	err := o.db.
		WithContext(ctx).
		Where("uuid = ?", uuid).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errOrder.ErrOrderNotFound)
		}
		return nil, errWrap.WrapError(errConst.ErrSQLError)
	}
	return &order, nil
}

func (o *OrderRepository) FindAllWithPagination(ctx context.Context, param *dto.OrderRequestParam) ([]models.Order, int64, error) {
	var (
		orders []models.Order
		sort   string
		total  int64
	)
	if param.SortColumn != nil {
		sort = fmt.Sprintf("%s %s", *param.SortColumn, *param.SortOrder)
	} else {
		sort = "created_at desc"
	}
	limit := param.Limit
	offset := (param.Page - 1) * limit
	err := o.db.
		WithContext(ctx).
		Limit(limit).
		Order(sort).
		Offset(offset).
		Find(&orders).
		Error

	if err != nil {
		return nil, 0, err
	}
	err = o.db.WithContext(ctx).Model(&models.Order{}).Count(&total).Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConst.ErrSQLError)
	}
	return orders, total, nil
}

func (o *OrderRepository) FindByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	var orders []models.Order
	err := o.db.WithContext(ctx).Where("user_id = ?", userID).Find(&orders).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errOrder.ErrOrderNotFound)
		}
		return nil, errWrap.WrapError(errConst.ErrSQLError)
	}
	return orders, nil
}

func (o *OrderRepository) incrementCode(ctx context.Context) (*string, error) {
	var (
		order  *models.Order
		result string
		today  = time.Now().Format("20060102")
	)
	err := o.db.WithContext(ctx).Order("id desc").First(&order).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errConst.ErrSQLError)
		}
	}
	if order.ID != 0 {
		orderCode := order.Code
		splitOrderName, _ := strconv.Atoi(orderCode[4:9])
		code := splitOrderName + 1
		result = fmt.Sprintf("ORD-%05d-%s", code, today)
	} else {
		result = fmt.Sprintf("ORD-%5d-%s", 1, today)
	}

	// ORD-00001-20250902
	return &result, nil
}

func (o *OrderRepository) Create(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error) {
	orderCode, err := o.incrementCode(ctx)
	if err != nil {
		return nil, err
	}
	result := &models.Order{
		UUID:   uuid.New(),
		Code:   *orderCode,
		UserID: order.UserID,
		Amount: order.Amount,
		Status: order.Status,
		Date:   order.Date,
		IsPaid: order.IsPaid,
	}
	err = tx.WithContext(ctx).Create(result).Error
	if err != nil {
		return nil, errWrap.WrapError(errConst.ErrSQLError)
	}
	return result, nil
}

func (o *OrderRepository) Update(ctx context.Context, tx *gorm.DB, uuid uuid.UUID, order *models.Order) error {
	err := tx.WithContext(ctx).Model(&models.Order{}).
		Where("uuid = ?", uuid).
		Updates(order).Error
	if err != nil {
		return errWrap.WrapError(errConst.ErrSQLError)
	}
	return nil
}

type IOrderRepository interface {
	FindAllWithPagination(ctx context.Context, param *dto.OrderRequestParam) ([]models.Order, int64, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Order, error)
	FindByUUID(ctx context.Context, uuid string) (*models.Order, error)
	Create(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error)
	Update(ctx context.Context, tx *gorm.DB, uuid uuid.UUID, order *models.Order) error
}
