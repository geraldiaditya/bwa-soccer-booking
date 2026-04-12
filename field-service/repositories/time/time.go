package repositories

import (
	"context"
	"errors"
	errWrap "field-service/common/error"
	errorConstants "field-service/constants/error"
	errTimes "field-service/constants/error/time"
	"field-service/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ITimeRepository interface {
	FindAll(context.Context) ([]models.Time, error)
	FindByUUID(context.Context, string) (*models.Time, error)
	FindByID(context.Context, int) (*models.Time, error)
	Create(context.Context, *models.Time) (*models.Time, error)
}

func NewTimeRepository(db *gorm.DB) ITimeRepository {
	return &TimeRepository{db: db}
}

type TimeRepository struct {
	db *gorm.DB
}

func (t *TimeRepository) FindAll(ctx context.Context) ([]models.Time, error) {
	var times []models.Time
	err := t.db.WithContext(ctx).Find(&times).Error
	if err != nil {
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return times, nil
}

func (t *TimeRepository) FindByUUID(ctx context.Context, uuid string) (*models.Time, error) {
	var time models.Time
	err := t.db.WithContext(ctx).Where("uuid = ?", uuid).First(&time).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errTimes.ErrTimeNotFound)
		}
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return &time, nil
}

func (t *TimeRepository) FindByID(ctx context.Context, id int) (*models.Time, error) {
	var time models.Time
	err := t.db.WithContext(ctx).Where("id = ?", id).First(&time).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errTimes.ErrTimeNotFound)
		}
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return &time, nil
}

func (t *TimeRepository) Create(ctx context.Context, time *models.Time) (*models.Time, error) {
	time.UUID = uuid.New()
	err := t.db.WithContext(ctx).Create(&time).Error
	if err != nil {
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return time, nil
}
