package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	errWrap "order-service/common/error"
	errConst "order-service/constants/error"
	errOrder "order-service/constants/error/order"
	"order-service/domain/dto"
	"order-service/domain/models"
	"strconv"
	"strings"
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
	if param.SortColumn != nil && *param.SortColumn != "" {
		col := strings.ToLower(*param.SortColumn)
		var ord string
		if param.SortOrder != nil {
			ord = strings.ToLower(*param.SortOrder)
		} else {
			ord = "desc"
		}
		
		allowedColumns := map[string]bool{
			"created_at": true, "amount": true, "status": true, "date": true,
		}
		allowedOrders := map[string]bool{"asc": true, "desc": true}

		if !allowedColumns[col] || !allowedOrders[ord] {
			return nil, 0, errWrap.WrapError(errConst.ErrRequestValidation)
		}
		sort = fmt.Sprintf("%s %s", col, ord)
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

// incrementCode acquires a pessimistic row lock (SELECT ... FOR UPDATE) to prevent
// duplicate codes under concurrent order creation within the same transaction.
func (o *OrderRepository) incrementCode(ctx context.Context, tx *gorm.DB) (*string, error) {
	var (
		order  models.Order
		result string
		today  = time.Now().Format("20060102")
	)
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Order("id desc").
		First(&order).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errConst.ErrSQLError)
		}
	}
	if order.ID != 0 {
		// Use SplitN to handle codes with more than 5 digits gracefully.
		parts := strings.SplitN(order.Code, "-", 3)
		var num int
		if len(parts) >= 2 {
			num, _ = strconv.Atoi(parts[1])
		}
		result = fmt.Sprintf("ORD-%05d-%s", num+1, today)
	} else {
		result = fmt.Sprintf("ORD-%05d-%s", 1, today)
	}
	return &result, nil
}

func (o *OrderRepository) Create(ctx context.Context, tx *gorm.DB, order *models.Order) (*models.Order, error) {
	orderCode, err := o.incrementCode(ctx, tx)
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
