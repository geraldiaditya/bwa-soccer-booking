package services

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"payment-service/constants"
	"payment-service/domain/models"
)

func TestMapPaymentResponseMapsAllFields(t *testing.T) {
	status := constants.Settlement
	transactionID := "transaction-123"
	vaNumber := "123456789"
	bank := "bca"
	invoiceLink := "https://example.com/invoice.pdf"
	acquirer := "gopay"
	description := "booking payment"
	paidAt := time.Date(2026, 5, 28, 1, 2, 3, 0, time.UTC)
	expiredAt := time.Date(2026, 5, 29, 1, 2, 3, 0, time.UTC)
	createdAt := time.Date(2026, 5, 27, 1, 2, 3, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 28, 2, 3, 4, 0, time.UTC)
	paymentUUID := uuid.New()
	orderID := uuid.New()

	result := mapPaymentResponse(models.Payment{
		ID:            10,
		UUID:          paymentUUID,
		OrderID:       orderID,
		Amount:        150000,
		Status:        &status,
		PaymentLink:   "https://example.com/pay",
		InvoiceLink:   &invoiceLink,
		VANumber:      &vaNumber,
		Bank:          &bank,
		Acquirer:      &acquirer,
		TransactionID: &transactionID,
		Description:   &description,
		PaidAt:        &paidAt,
		ExpiredAt:     &expiredAt,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	})

	if result.UUID != paymentUUID {
		t.Fatalf("expected UUID %s, got %s", paymentUUID, result.UUID)
	}
	if result.OrderID != orderID {
		t.Fatalf("expected order ID %s, got %s", orderID, result.OrderID)
	}
	if result.Amount != 150000 {
		t.Fatalf("expected amount 150000, got %v", result.Amount)
	}
	if result.Status != constants.SettlementString {
		t.Fatalf("expected status %s, got %s", constants.SettlementString, result.Status)
	}
	if result.PaymentLink != "https://example.com/pay" {
		t.Fatalf("expected payment link to be mapped")
	}
	assertStringPtrEqual(t, "transaction id", &transactionID, result.TransactionId)
	assertTimePtrEqual(t, "paid at", &paidAt, result.PaidAt)
	assertStringPtrEqual(t, "va number", &vaNumber, result.VANumber)
	assertStringPtrEqual(t, "bank", &bank, result.Bank)
	assertStringPtrEqual(t, "invoice link", &invoiceLink, result.InvoiceLink)
	assertStringPtrEqual(t, "acquirer", &acquirer, result.Acquirer)
	assertStringPtrEqual(t, "description", &description, result.Description)
	assertTimePtrEqual(t, "created at", &createdAt, result.CreatedAt)
	assertTimePtrEqual(t, "updated at", &updatedAt, result.UpdatedAt)
}

func TestMapPaymentResponseKeepsNilOptionalFields(t *testing.T) {
	status := constants.Pending

	result := mapPaymentResponse(models.Payment{
		Status: &status,
	})

	if result.TransactionId != nil {
		t.Fatalf("expected transaction id to be nil")
	}
	if result.PaidAt != nil {
		t.Fatalf("expected paid at to be nil")
	}
	if result.VANumber != nil {
		t.Fatalf("expected va number to be nil")
	}
	if result.Bank != nil {
		t.Fatalf("expected bank to be nil")
	}
	if result.InvoiceLink != nil {
		t.Fatalf("expected invoice link to be nil")
	}
	if result.Acquirer != nil {
		t.Fatalf("expected acquirer to be nil")
	}
	if result.Description != nil {
		t.Fatalf("expected description to be nil")
	}
	if result.CreatedAt != nil {
		t.Fatalf("expected created at to be nil")
	}
	if result.UpdatedAt != nil {
		t.Fatalf("expected updated at to be nil")
	}
}

func TestMapPaymentResponsesMapsSlice(t *testing.T) {
	pending := constants.Pending
	settlement := constants.Settlement

	result := mapPaymentResponses([]models.Payment{
		{Status: &pending, Amount: 1},
		{Status: &settlement, Amount: 2},
	})

	if len(result) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(result))
	}
	if result[0].Status != constants.PendingString || result[0].Amount != 1 {
		t.Fatalf("expected first payment to be mapped")
	}
	if result[1].Status != constants.SettlementString || result[1].Amount != 2 {
		t.Fatalf("expected second payment to be mapped")
	}
}

func assertStringPtrEqual(t *testing.T, field string, expected, actual *string) {
	t.Helper()
	if expected == nil || actual == nil || *expected != *actual {
		t.Fatalf("expected %s %v, got %v", field, expected, actual)
	}
}

func assertTimePtrEqual(t *testing.T, field string, expected, actual *time.Time) {
	t.Helper()
	if expected == nil || actual == nil || !expected.Equal(*actual) {
		t.Fatalf("expected %s %v, got %v", field, expected, actual)
	}
}
