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

func TestFieldSortReturnsRequestedSort(t *testing.T) {
	sortColumn := "name"
	sortOrder := "asc"

	result := fieldSort(&dto.FieldRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "name asc", result)
}
