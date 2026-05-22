package repositories

import (
	"field-service/domain/dto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldSortReturnsDefaultSort(t *testing.T) {
	result := fieldSort(&dto.FieldRequestParam{})

	assert.Equal(t, defaultFieldSort, result)
}

func TestFieldSortReturnsDefaultSortWhenSortParamIsPartial(t *testing.T) {
	sortColumn := "name"

	result := fieldSort(&dto.FieldRequestParam{
		SortColumn: &sortColumn,
	})

	assert.Equal(t, defaultFieldSort, result)
}

func TestFieldSortReturnsRequestedSort(t *testing.T) {
	sortColumn := "name"
	sortOrder := "asc"

	result := fieldSort(&dto.FieldRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "name asc", result)
}

func TestFieldSortReturnsDefaultSortWhenColumnIsNotAllowed(t *testing.T) {
	sortColumn := "name; DROP TABLE fields; --"
	sortOrder := "asc"

	result := fieldSort(&dto.FieldRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, defaultFieldSort, result)
}

func TestFieldSortDefaultsToDescWhenOrderIsNotAsc(t *testing.T) {
	sortColumn := "name"
	sortOrder := "desc; DROP TABLE fields; --"

	result := fieldSort(&dto.FieldRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "name desc", result)
}

func TestFieldPaginationKeepsOffsetNonNegative(t *testing.T) {
	limit, offset := fieldPagination(&dto.FieldRequestParam{
		Page:  0,
		Limit: 10,
	})

	assert.Equal(t, 10, limit)
	assert.Equal(t, 0, offset)
}
