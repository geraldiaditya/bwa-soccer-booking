package repositories

import (
	"payment-service/domain/dto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaymentSortReturnsDefaultSort(t *testing.T) {
	result := paymentSort(&dto.PaymentRequestParam{})

	assert.Equal(t, defaultPaymentSort, result)
}

func TestPaymentSortReturnsDefaultSortWhenSortParamIsPartial(t *testing.T) {
	sortColumn := "amount"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
	})

	assert.Equal(t, defaultPaymentSort, result)
}

func TestPaymentSortReturnsRequestedSort(t *testing.T) {
	sortColumn := "amount"
	sortOrder := "asc"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "amount asc", result)
}

func TestPaymentSortReturnsDefaultSortWhenColumnIsNotAllowed(t *testing.T) {
	sortColumn := "amount; DROP TABLE payments; --"
	sortOrder := "asc"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, defaultPaymentSort, result)
}

func TestPaymentSortDefaultsToDescWhenOrderIsNotAsc(t *testing.T) {
	sortColumn := "amount"
	sortOrder := "desc; DROP TABLE payments; --"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assert.Equal(t, "amount desc", result)
}

func TestPaymentPaginationKeepsOffsetNonNegative(t *testing.T) {
	limit, offset := paymentPagination(&dto.PaymentRequestParam{
		Page:  0,
		Limit: 10,
	})

	assert.Equal(t, 10, limit)
	assert.Equal(t, 0, offset)
}
