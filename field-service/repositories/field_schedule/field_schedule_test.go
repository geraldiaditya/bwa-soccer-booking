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

func TestFieldScheduleSortReturnsRequestedSort(t *testing.T) {
	sortColumn := "date"
	sortOrder := "desc"

	result := fieldScheduleSort(&dto.FieldScheduleRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "date desc", result)
}
