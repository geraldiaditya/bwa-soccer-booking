package repositories

import (
	"field-service/domain/dto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldScheduleSortReturnsDefaultSort(t *testing.T) {
	result := fieldScheduleSort(&dto.FieldScheduleRequestParam{})

	assert.Equal(t, defaultFieldScheduleSort, result)
}

func TestFieldScheduleSortReturnsDefaultSortWhenSortParamIsPartial(t *testing.T) {
	sortColumn := "date"

	result := fieldScheduleSort(&dto.FieldScheduleRequestParam{
		SortColumn: &sortColumn,
	})

	assert.Equal(t, defaultFieldScheduleSort, result)
}

func TestFieldScheduleSortReturnsRequestedSort(t *testing.T) {
	sortColumn := "date"
	sortOrder := "desc"

	result := fieldScheduleSort(&dto.FieldScheduleRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "date desc", result)
}

func TestFieldScheduleSortReturnsDefaultSortWhenColumnIsNotAllowed(t *testing.T) {
	sortColumn := "date; DROP TABLE field_schedules; --"
	sortOrder := "asc"

	result := fieldScheduleSort(&dto.FieldScheduleRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, defaultFieldScheduleSort, result)
}

func TestFieldScheduleSortDefaultsToDescWhenOrderIsNotAsc(t *testing.T) {
	sortColumn := "date"
	sortOrder := "desc; DROP TABLE field_schedules; --"

	result := fieldScheduleSort(&dto.FieldScheduleRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "date desc", result)
}

func TestFieldSchedulePaginationKeepsOffsetNonNegative(t *testing.T) {
	limit, offset := fieldSchedulePagination(&dto.FieldScheduleRequestParam{
		Page:  0,
		Limit: 10,
	})

	assert.Equal(t, 10, limit)
	assert.Equal(t, 0, offset)
}
