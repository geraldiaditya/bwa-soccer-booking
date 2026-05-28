package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"payment-service/domain/dto"
	"payment-service/domain/models"
)

type fakeGCSClient struct {
	filename string
	data     []byte
}

func (f *fakeGCSClient) UploadFile(_ context.Context, filename string, data []byte) (string, error) {
	f.filename = filename
	f.data = data
	return "https://storage.googleapis.com/test-bucket/" + filename, nil
}

func TestBuildInvoiceRequest(t *testing.T) {
	bank := "bca"
	vaNumber := "1234567890"
	description := "Soccer booking"
	payment := &models.Payment{
		Amount:      150000,
		Bank:        &bank,
		VANumber:    &vaNumber,
		Description: &description,
	}
	webhook := &dto.Webhook{PaymentType: "bank_transfer"}
	paidAt := time.Date(2026, time.October, 5, 10, 30, 0, 0, time.UTC)

	got := BuildInvoiceRequest(payment, webhook, paidAt, "INV/2026-10-05/ORD/order-payment")

	if got.InvoiceNumber != "INV/2026-10-05/ORD/order-payment" {
		t.Fatalf("InvoiceNumber = %q", got.InvoiceNumber)
	}
	detail := got.Data.PaymentDetail
	if detail.BankName != "BCA" {
		t.Fatalf("BankName = %q, want BCA", detail.BankName)
	}
	if detail.PaymentMethod != "bank_transfer" {
		t.Fatalf("PaymentMethod = %q, want bank_transfer", detail.PaymentMethod)
	}
	if detail.VANumber != vaNumber {
		t.Fatalf("VANumber = %q, want %q", detail.VANumber, vaNumber)
	}
	if detail.Date != "05 Oktober 2026" {
		t.Fatalf("Date = %q, want %q", detail.Date, "05 Oktober 2026")
	}
	if !detail.IsPaid {
		t.Fatal("IsPaid = false, want true")
	}
	if got.Data.Total != "Rp. 150.000" {
		t.Fatalf("Total = %q, want %q", got.Data.Total, "Rp. 150.000")
	}
	if len(got.Data.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(got.Data.Items))
	}
	if got.Data.Items[0].Description != description {
		t.Fatalf("Item description = %q, want %q", got.Data.Items[0].Description, description)
	}
	if got.Data.Items[0].Price != "Rp. 150.000" {
		t.Fatalf("Item price = %q, want %q", got.Data.Items[0].Price, "Rp. 150.000")
	}
}

func TestBuildInvoiceRequestHandlesNilOptionalFields(t *testing.T) {
	payment := &models.Payment{
		Amount: 150000,
	}
	webhook := &dto.Webhook{PaymentType: "bank_transfer"}
	paidAt := time.Date(2026, time.October, 5, 10, 30, 0, 0, time.UTC)

	got := BuildInvoiceRequest(payment, webhook, paidAt, "INV/2026-10-05/ORD/order-payment")

	detail := got.Data.PaymentDetail
	if detail.BankName != "" {
		t.Fatalf("BankName = %q, want empty", detail.BankName)
	}
	if detail.VANumber != "" {
		t.Fatalf("VANumber = %q, want empty", detail.VANumber)
	}
	if got.Data.Items[0].Description != "" {
		t.Fatalf("Item description = %q, want empty", got.Data.Items[0].Description)
	}
}

func TestBuildInvoiceNumberUsesUUIDSegments(t *testing.T) {
	orderID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	paymentID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	now := time.Date(2026, time.May, 28, 0, 0, 0, 0, time.UTC)

	got := BuildInvoiceNumber(now, orderID, paymentID)
	want := "INV/2026-05-28/ORD/11111111-aaaaaaaa"
	if got != want {
		t.Fatalf("BuildInvoiceNumber() = %q, want %q", got, want)
	}
}

func TestInvoiceFilename(t *testing.T) {
	got := InvoiceFilename("INV/2026-05-28/ORD/11111111-aaaaaaaa")
	want := "inv-2026-05-28-ord-11111111-aaaaaaaa.pdf"
	if got != want {
		t.Fatalf("InvoiceFilename() = %q, want %q", got, want)
	}
}

func TestInvoiceGeneratorGenerateAndUpload(t *testing.T) {
	gcsClient := &fakeGCSClient{}
	generator := NewInvoiceGenerator(gcsClient)
	generator.renderPDF = func(req *dto.InvoiceRequest) ([]byte, error) {
		if req.InvoiceNumber != "INV/2026-05-28/ORD/11111111-aaaaaaaa" {
			t.Fatalf("InvoiceNumber = %q", req.InvoiceNumber)
		}
		return []byte("pdf-bytes"), nil
	}

	url, err := generator.GenerateAndUpload(context.Background(), &dto.InvoiceRequest{
		InvoiceNumber: "INV/2026-05-28/ORD/11111111-aaaaaaaa",
	})
	if err != nil {
		t.Fatalf("GenerateAndUpload() error = %v", err)
	}
	if gcsClient.filename != "inv-2026-05-28-ord-11111111-aaaaaaaa.pdf" {
		t.Fatalf("uploaded filename = %q", gcsClient.filename)
	}
	if string(gcsClient.data) != "pdf-bytes" {
		t.Fatalf("uploaded data = %q", string(gcsClient.data))
	}
	if url != "https://storage.googleapis.com/test-bucket/inv-2026-05-28-ord-11111111-aaaaaaaa.pdf" {
		t.Fatalf("url = %q", url)
	}
}
