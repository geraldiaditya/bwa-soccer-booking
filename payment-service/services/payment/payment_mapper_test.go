package services

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"payment-service/constants"
	"payment-service/domain/dto"
	"payment-service/domain/models"
)

func TestMapPaymentResponseMapsAllFields(t *testing.T) {
	status := constants.Settlement
	transactionID := "transaction-1"
	vaNumber := "123456789"
	bank := "bca"
	invoiceLink := "https://example.com/invoice.pdf"
	acquirer := "gopay"
	description := "booking payment"
	paidAt := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 28, 11, 0, 0, 0, time.UTC)
	payment := models.Payment{
		UUID:          uuid.New(),
		OrderID:       uuid.New(),
		Amount:        250000,
		Status:        &status,
		PaymentLink:   "https://example.com/pay",
		TransactionID: &transactionID,
		PaidAt:        &paidAt,
		VANumber:      &vaNumber,
		Bank:          &bank,
		InvoiceLink:   &invoiceLink,
		Acquirer:      &acquirer,
		Description:   &description,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}

	result := mapPaymentResponse(payment)

	expected := dto.PaymentResponse{
		UUID:          payment.UUID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        constants.SettlementString,
		PaymentLink:   payment.PaymentLink,
		TransactionId: &transactionID,
		PaidAt:        &paidAt,
		VANumber:      &vaNumber,
		Bank:          &bank,
		InvoiceLink:   &invoiceLink,
		Acquirer:      &acquirer,
		Description:   &description,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expected %+v, got %+v", expected, result)
	}
}

func TestMapPaymentResponseHandlesNilOptionalFields(t *testing.T) {
	payment := models.Payment{
		UUID:        uuid.New(),
		OrderID:     uuid.New(),
		Amount:      100000,
		PaymentLink: "https://example.com/pay",
	}

	result := mapPaymentResponse(payment)

	if result.TransactionId != nil {
		t.Fatalf("expected nil transaction id, got %v", *result.TransactionId)
	}
	if result.PaidAt != nil {
		t.Fatalf("expected nil paid at, got %v", *result.PaidAt)
	}
	if result.VANumber != nil {
		t.Fatalf("expected nil va number, got %v", *result.VANumber)
	}
	if result.Bank != nil {
		t.Fatalf("expected nil bank, got %v", *result.Bank)
	}
	if result.InvoiceLink != nil {
		t.Fatalf("expected nil invoice link, got %v", *result.InvoiceLink)
	}
	if result.Acquirer != nil {
		t.Fatalf("expected nil acquirer, got %v", *result.Acquirer)
	}
	if result.Description != nil {
		t.Fatalf("expected nil description, got %v", *result.Description)
	}
	if result.CreatedAt != nil {
		t.Fatalf("expected nil created at, got %v", *result.CreatedAt)
	}
	if result.UpdatedAt != nil {
		t.Fatalf("expected nil updated at, got %v", *result.UpdatedAt)
	}
	if result.Status != constants.PaymentStatusString("") {
		t.Fatalf("expected empty status, got %q", result.Status)
	}
}

func TestMapPaymentResponsesMapsSlice(t *testing.T) {
	status := constants.Pending
	payments := []models.Payment{
		{UUID: uuid.New(), OrderID: uuid.New(), Amount: 10, Status: &status, PaymentLink: "https://example.com/one"},
		{UUID: uuid.New(), OrderID: uuid.New(), Amount: 20, Status: &status, PaymentLink: "https://example.com/two"},
	}

	result := mapPaymentResponses(payments)

	if len(result) != len(payments) {
		t.Fatalf("expected %d responses, got %d", len(payments), len(result))
	}
	for i := range payments {
		if result[i] != mapPaymentResponse(payments[i]) {
			t.Fatalf("unexpected response at index %d: %+v", i, result[i])
		}
	}
}
