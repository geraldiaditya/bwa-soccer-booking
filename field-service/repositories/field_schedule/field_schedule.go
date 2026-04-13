package repositories

import (
	"context"
	"errors"
	errWrap "field-service/common/error"
	"field-service/constants"
	errorConstants "field-service/constants/error"
	errFieldSchedule "field-service/constants/error/field_schedule"
	"field-service/domain/dto"
	"field-service/domain/models"
	"fmt"
	"gorm.io/gorm"
)

type IFieldScheduleRepository interface {
	FindAllWithPagination(context.Context, *dto.FieldScheduleRequestParam) ([]models.FieldSchedule, int64, error)
	FindAllByFieldIdAndDate(context.Context, int, string) ([]models.FieldSchedule, error)
	FindByUUID(context.Context, string) (*models.FieldSchedule, error)
	FindByDateAndTimeId(context.Context, string, int, int) (*models.FieldSchedule, error)
	Create(context.Context, []models.FieldSchedule) error
	Update(context.Context, string, *models.FieldSchedule) (*models.FieldSchedule, error)
	UpdateStatus(context.Context, constants.FieldScheduleStatus, string) error
	Delete(context.Context, string) error
}

func NewFieldScheduleRepository(db *gorm.DB) IFieldScheduleRepository {
	return &FieldScheduleRepository{db: db}
}

type FieldScheduleRepository struct {
	db *gorm.DB
}

func (f *FieldScheduleRepository) FindAllWithPagination(
	ctx context.Context,
	param *dto.FieldScheduleRequestParam,
) ([]models.FieldSchedule, int64, error) {
	var (
		fieldSchedules []models.FieldSchedule
		sort           string
		total          int64
	)
	sort = "created_at desc"
	if param.SortColumn != nil {
		sort = fmt.Sprintf("%s %s", *param.SortColumn, *param.SortOrder)
	}

	limit := param.Limit
	offset := (param.Page - 1) * limit
	err := f.db.
		WithContext(ctx).
		Preload("Field").
		Preload("Time").
		Limit(limit).
		Offset(offset).
		Order(sort).
		Find(&fieldSchedules).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	err = f.db.
		WithContext(ctx).
		Model(&fieldSchedules).
		Count(&total).
		Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return fieldSchedules, total, nil
}

func (f *FieldScheduleRepository) FindAllByFieldIdAndDate(ctx context.Context, fieldId int, date string) ([]models.FieldSchedule, error) {
	var fieldSchedules []models.FieldSchedule
	err := f.db.
		WithContext(ctx).
		Preload("Field").
		Preload("Time").
		Where("field_id = ?", fieldId).
		Where("date = ?", date).
		Joins("LEFT JOIN times ON field_schedules.time_id = times.id").
		Order("times.start_time asc").
		Find(&fieldSchedules).
		Error
	if err != nil {
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return fieldSchedules, nil
}

func (f *FieldScheduleRepository) FindByUUID(ctx context.Context, uuid string) (*models.FieldSchedule, error) {
	var fieldSchedule models.FieldSchedule
	err := f.db.
		WithContext(ctx).
		Preload("Field").
		Preload("Time").
		Where("uuid = ?", uuid).
		First(&fieldSchedule).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errFieldSchedule.ErrFieldScheduleNotFound)
		}
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return &fieldSchedule, nil
}
func (f *FieldScheduleRepository) FindByDateAndTimeId(ctx context.Context, date string, timeId int, fieldId int) (*models.FieldSchedule, error) {
	var fieldSchedule models.FieldSchedule
	err := f.db.
		WithContext(ctx).
		Where("date = ?", date).
		Where("time_id = ?", timeId).
		Where("field_id = ?", fieldId).
		First(&fieldSchedule).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return &fieldSchedule, nil
}

func (f *FieldScheduleRepository) Create(ctx context.Context, req []models.FieldSchedule) error {
	err := f.db.WithContext(ctx).Create(&req).Error
	if err != nil {
		return errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return nil
}

func (f *FieldScheduleRepository) Update(ctx context.Context, uuid string, req *models.FieldSchedule) (*models.FieldSchedule, error) {
	fieldSchedule, err := f.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	fieldSchedule.Date = req.Date

	err = f.db.WithContext(ctx).Save(&fieldSchedule).Error
	if err != nil {
		return nil, errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return fieldSchedule, nil
}
func (f *FieldScheduleRepository) UpdateStatus(ctx context.Context, status constants.FieldScheduleStatus, uuid string) error {
	fieldSchedule, err := f.FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	fieldSchedule.Status = status
	err = f.db.WithContext(ctx).Save(&fieldSchedule).Error
	if err != nil {
		return errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return nil
}

func (f *FieldScheduleRepository) Delete(ctx context.Context, uuid string) error {
	err := f.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&models.FieldSchedule{}).Error
	if err != nil {
		return errWrap.WrapError(errorConstants.ErrSQLError)
	}
	return nil
}
