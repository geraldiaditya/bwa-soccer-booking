package services

import (
	"context"
	"field-service/common/utils"
	"field-service/constants"
	errFieldSchedule "field-service/constants/error/field_schedule"
	"field-service/domain/dto"
	"field-service/domain/models"
	"field-service/repositories"
	"time"

	"github.com/google/uuid"
)

type IFieldScheduleService interface {
	GetAllWithPagination(context.Context, *dto.FieldScheduleRequestParam) (*utils.PaginationResult, error)
	GetAllByFieldIdAndDate(context.Context, string, string) ([]dto.FieldScheduleForBookingResponse, error)
	GetByUUID(context.Context, string) (*dto.FieldScheduleResponse, error)
	GenerateScheduleForOneMonth(context.Context, *dto.GenerateFieldScheduleForOneMonthRequest) error
	Create(context.Context, *dto.FieldScheduleRequest) error
	Update(context.Context, string, *dto.UpdateFieldScheduleRequest) (*dto.FieldScheduleResponse, error)
	UpdateStatus(context.Context, *dto.UpdateStatusFieldScheduleRequest) error
	Delete(context.Context, string) error
}

func NewFieldScheduleService(repository repositories.IRepositoryRegistry) IFieldScheduleService {
	return &FieldScheduleService{repository: repository}
}

type FieldScheduleService struct {
	repository repositories.IRepositoryRegistry
}

func (f *FieldScheduleService) GetAllWithPagination(
	ctx context.Context,
	param *dto.FieldScheduleRequestParam,
) (*utils.PaginationResult, error) {
	fieldSchedules, total, err := f.repository.GetFieldSchedule().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}
	pagination := utils.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  toFieldScheduleResponses(fieldSchedules),
	}
	response := utils.GeneratePagination(pagination)
	return &response, nil
}

func (f *FieldScheduleService) GetAllByFieldIdAndDate(
	ctx context.Context, uuid string, date string,
) ([]dto.FieldScheduleForBookingResponse, error) {
	field, err := f.repository.GetField().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	fieldSchedules, err := f.repository.GetFieldSchedule().FindAllByFieldIdAndDate(ctx, int(field.ID), date)
	if err != nil {
		return nil, err
	}
	return toFieldScheduleForBookingResponses(fieldSchedules), nil
}

func (f *FieldScheduleService) GetByUUID(ctx context.Context, uuid string) (*dto.FieldScheduleResponse, error) {
	fieldSchedule, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	response := toFieldScheduleResponse(*fieldSchedule)
	return &response, nil
}

func (f *FieldScheduleService) GenerateScheduleForOneMonth(ctx context.Context, request *dto.GenerateFieldScheduleForOneMonthRequest) error {
	return f.repository.WithTransaction(ctx, func(repository repositories.IRepositoryRegistry) error {
		field, err := repository.GetField().FindByUUID(ctx, request.FieldID)
		if err != nil {
			return err
		}
		times, err := repository.GetTime().FindAll(ctx)
		if err != nil {
			return err
		}
		numberOfDays := 30
		fieldSchedules := make([]models.FieldSchedule, 0, numberOfDays)
		now := time.Now().Add(time.Duration(1) * 24 * time.Hour)
		for i := 0; i < numberOfDays; i++ {
			currentDate := now.AddDate(0, 0, i)
			for _, item := range times {
				schedule, err := repository.GetFieldSchedule().
					FindByDateAndTimeId(
						ctx, currentDate.Format(time.DateOnly), int(item.ID), int(field.ID),
					)
				if err != nil {
					return err
				}

				if schedule != nil {
					return errFieldSchedule.ErrFieldScheduleIsExist
				}
				fieldSchedules = append(fieldSchedules, models.FieldSchedule{
					UUID:    uuid.New(),
					FieldID: field.ID,
					TimeID:  item.ID,
					Date:    currentDate,
					Status:  constants.Available,
				})
			}
		}
		err = repository.GetFieldSchedule().Create(ctx, fieldSchedules)
		if err != nil {
			return err
		}
		return nil
	})
}

func (f *FieldScheduleService) Create(ctx context.Context, request *dto.FieldScheduleRequest) error {
	return f.repository.WithTransaction(ctx, func(repository repositories.IRepositoryRegistry) error {
		field, err := repository.GetField().FindByUUID(ctx, request.FieldID)
		if err != nil {
			return err
		}
		fieldSchedules := make([]models.FieldSchedule, 0, len(request.TimeIDs))
		dateParsed, _ := time.Parse(time.DateOnly, request.Date)
		for _, timeID := range request.TimeIDs {
			scheduleTime, err := repository.GetTime().FindByUUID(ctx, timeID)
			if err != nil {
				return err
			}
			schedule, err := repository.GetFieldSchedule().
				FindByDateAndTimeId(ctx, request.Date, int(scheduleTime.ID), int(field.ID))
			if err != nil {
				return err
			}
			if schedule != nil {
				return errFieldSchedule.ErrFieldScheduleIsExist
			}
			fieldSchedules = append(fieldSchedules, models.FieldSchedule{
				UUID:    uuid.New(),
				FieldID: field.ID,
				TimeID:  scheduleTime.ID,
				Date:    dateParsed,
				Status:  constants.Available,
			})
		}
		err = repository.GetFieldSchedule().Create(ctx, fieldSchedules)
		if err != nil {
			return err
		}
		return nil
	})
}

func (f *FieldScheduleService) Update(
	ctx context.Context,
	uuid string,
	request *dto.UpdateFieldScheduleRequest,
) (*dto.FieldScheduleResponse, error) {
	var response dto.FieldScheduleResponse
	err := f.repository.WithTransaction(ctx, func(repository repositories.IRepositoryRegistry) error {
		fieldSchedule, err := repository.GetFieldSchedule().FindByUUID(ctx, uuid)
		if err != nil {
			return err
		}
		scheduleTime, err := repository.GetTime().FindByUUID(ctx, request.TimeID)
		if err != nil {
			return err
		}
		isTimeExist, err := repository.GetFieldSchedule().FindByDateAndTimeId(
			ctx,
			request.Date,
			int(scheduleTime.ID),
			int(fieldSchedule.Field.ID),
		)
		if err != nil {
			return err
		}
		if isTimeExist != nil && request.Date != fieldSchedule.Date.Format(time.DateOnly) {
			checkDate, err := repository.GetFieldSchedule().FindByDateAndTimeId(
				ctx,
				request.Date,
				int(scheduleTime.ID),
				int(fieldSchedule.Field.ID),
			)
			if err != nil {
				return err
			}
			if checkDate != nil {
				return errFieldSchedule.ErrFieldScheduleIsExist
			}
		}
		dateParsed, _ := time.Parse(time.DateOnly, request.Date)
		fieldResult, err := repository.GetFieldSchedule().Update(ctx, uuid, &models.FieldSchedule{
			Date:   dateParsed,
			TimeID: scheduleTime.ID,
		})
		if err != nil {
			return err
		}
		response = toFieldScheduleResponseWithTime(*fieldResult, *scheduleTime)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *FieldScheduleService) UpdateStatus(
	ctx context.Context,
	request *dto.UpdateStatusFieldScheduleRequest,
) error {
	return f.repository.WithTransaction(ctx, func(repository repositories.IRepositoryRegistry) error {
		for _, item := range request.FieldScheduleIDs {
			_, err := repository.GetFieldSchedule().FindByUUID(ctx, item)
			if err != nil {
				return err
			}
			err = repository.GetFieldSchedule().UpdateStatus(ctx, constants.Booked, item)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (f *FieldScheduleService) Delete(ctx context.Context, uuid string) error {
	_, err := f.repository.GetFieldSchedule().FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	err = f.repository.GetFieldSchedule().Delete(ctx, uuid)
	if err != nil {
		return err
	}
	return nil
}
