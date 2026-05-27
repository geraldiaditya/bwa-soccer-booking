package repositories

import (
	"payment-service/domain/dto"
	"testing"
)

func TestPaymentSortReturnsDefaultSort(t *testing.T) {
	result := paymentSort(&dto.PaymentRequestParam{})

	assertEqual(t, defaultPaymentSort, result)
}

func TestPaymentSortReturnsDefaultSortWhenSortParamIsPartial(t *testing.T) {
	sortColumn := "amount"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
	})

	assertEqual(t, defaultPaymentSort, result)
}

func TestPaymentSortReturnsRequestedSort(t *testing.T) {
	sortColumn := "amount"
	sortOrder := "asc"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assertEqual(t, "amount asc", result)
}

func TestPaymentSortReturnsDescWhenOrderIsDesc(t *testing.T) {
	sortColumn := "amount"
	sortOrder := "desc"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assertEqual(t, "amount desc", result)
}

func TestPaymentSortNormalizesColumnCase(t *testing.T) {
	sortColumn := "Amount"
	sortOrder := "asc"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assertEqual(t, "amount asc", result)
}

func TestPaymentSortReturnsDefaultSortWhenColumnIsNotAllowed(t *testing.T) {
	sortColumn := "amount; DROP TABLE payments; --"
	sortOrder := "asc"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assertEqual(t, defaultPaymentSort, result)
}

func TestPaymentSortDefaultsToDescWhenOrderIsNotAsc(t *testing.T) {
	sortColumn := "amount"
	sortOrder := "desc; DROP TABLE payments; --"

	result := paymentSort(&dto.PaymentRequestParam{
		SortColumn: &sortColumn,
		SortOrder:  &sortOrder,
	})

	assertEqual(t, "amount desc", result)
}

func TestPaymentPaginationKeepsOffsetNonNegative(t *testing.T) {
	limit, offset := paymentPagination(&dto.PaymentRequestParam{
		Page:  0,
		Limit: 10,
	})

	assertEqual(t, 10, limit)
	assertEqual(t, 0, offset)
}

func TestPaymentPaginationDefaultsInvalidLimit(t *testing.T) {
	limit, offset := paymentPagination(&dto.PaymentRequestParam{
		Page:  2,
		Limit: 0,
	})

	assertEqual(t, defaultPaymentLimit, limit)
	assertEqual(t, defaultPaymentLimit, offset)
}

func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
