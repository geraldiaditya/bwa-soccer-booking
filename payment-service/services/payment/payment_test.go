package services

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	clients "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	config2 "payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
)

type fakeKafkaRegistry struct {
	producer *fakeKafkaProducer
}

func (f *fakeKafkaRegistry) GetKafkaProducer() kafka.IKafka {
	return f.producer
}

type fakeKafkaProducer struct {
	messages [][]byte
}

func (f *fakeKafkaProducer) ProduceMessage(_ string, data []byte) error {
	f.messages = append(f.messages, data)
	return nil
}

type fakeGCSClient struct{}

func (f *fakeGCSClient) UploadFile(context.Context, string, []byte) (string, error) {
	return "https://example.com/invoice.pdf", nil
}

type fakeMidtransClient struct{}

func (f *fakeMidtransClient) CreatePaymentLink(*dto.PaymentRequest) (*clients.MidTransData, error) {
	return &clients.MidTransData{RedirectURL: "https://example.com/pay"}, nil
}

func newTestPaymentService(t *testing.T) (*PaymentService, *gorm.DB, *fakeKafkaProducer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&models.Payment{}, &models.PaymentHistory{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	producer := &fakeKafkaProducer{}
	service := NewPaymentService(
		repositories.NewRepositoryRegistry(db),
		&fakeGCSClient{},
		&fakeKafkaRegistry{producer: producer},
		&fakeMidtransClient{},
	).(*PaymentService)
	config2.Config.Midtrans.ServerKey = "server-key"
	config2.Config.Kafka.Topic = "payment-events"
	return service, db, producer
}

func createPayment(t *testing.T, db *gorm.DB, status constants.PaymentStatus) models.Payment {
	t.Helper()
	description := "booking"
	expiredAt := time.Now().Add(time.Hour)
	payment := models.Payment{
		UUID:        uuid.New(),
		OrderID:     uuid.New(),
		Amount:      100000,
		Status:      &status,
		PaymentLink: "https://example.com/pay",
		Description: &description,
		ExpiredAt:   &expiredAt,
	}
	if err := db.Create(&payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}
	return payment
}

func signedWebhook(service *PaymentService, payment models.Payment, status constants.PaymentStatusString) *dto.Webhook {
	webhook := &dto.Webhook{
		OrderId:           payment.OrderID,
		TransactionId:     uuid.NewString(),
		TransactionStatus: status,
		StatusCode:        "200",
		GrossAmount:       "100000.00",
		PaymentType:       "bank_transfer",
	}
	webhook.SignatureKey = service.webhookSignature(
		webhook.OrderId.String(),
		webhook.StatusCode,
		webhook.GrossAmount,
		config2.Config.Midtrans.ServerKey,
	)
	return webhook
}

func TestWebhookRejectsInvalidSignature(t *testing.T) {
	service, db, producer := newTestPaymentService(t)
	payment := createPayment(t, db, constants.Initial)
	webhook := signedWebhook(service, payment, constants.PendingString)
	webhook.SignatureKey = "invalid"

	err := service.Webhook(context.Background(), webhook)

	if !errors.Is(err, errPayment.ErrWebhookInvalidSignature) {
		t.Fatalf("expected invalid signature error, got %v", err)
	}
	var stored models.Payment
	if err = db.First(&stored, payment.ID).Error; err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if *stored.Status != constants.Initial {
		t.Fatalf("expected unchanged status initial, got %v", *stored.Status)
	}
	if len(producer.messages) != 0 {
		t.Fatalf("expected no kafka messages, got %d", len(producer.messages))
	}
}

func TestWebhookRejectsUnsupportedPayload(t *testing.T) {
	service, _, _ := newTestPaymentService(t)
	tests := []struct {
		name    string
		webhook *dto.Webhook
	}{
		{
			name:    "nil webhook",
			webhook: nil,
		},
		{
			name: "missing order id",
			webhook: &dto.Webhook{
				TransactionId:     uuid.NewString(),
				TransactionStatus: constants.PendingString,
				StatusCode:        "200",
				GrossAmount:       "100000.00",
				SignatureKey:      "signature",
			},
		},
		{
			name: "unsupported status",
			webhook: &dto.Webhook{
				OrderId:           uuid.New(),
				TransactionId:     uuid.NewString(),
				TransactionStatus: "deny",
				StatusCode:        "200",
				GrossAmount:       "100000.00",
				SignatureKey:      "signature",
			},
		},
		{
			name: "missing signature",
			webhook: &dto.Webhook{
				OrderId:           uuid.New(),
				TransactionId:     uuid.NewString(),
				TransactionStatus: constants.PendingString,
				StatusCode:        "200",
				GrossAmount:       "100000.00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Webhook(context.Background(), tt.webhook)

			if !errors.Is(err, errPayment.ErrWebhookInvalidPayload) {
				t.Fatalf("expected invalid payload error, got %v", err)
			}
		})
	}
}

func TestWebhookHandlesPendingWithoutVANumbers(t *testing.T) {
	service, db, producer := newTestPaymentService(t)
	payment := createPayment(t, db, constants.Initial)
	webhook := signedWebhook(service, payment, constants.PendingString)

	err := service.Webhook(context.Background(), webhook)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	var stored models.Payment
	if err = db.First(&stored, payment.ID).Error; err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if *stored.Status != constants.Pending {
		t.Fatalf("expected pending status, got %v", *stored.Status)
	}
	if stored.VANumber != nil || stored.Bank != nil {
		t.Fatalf("expected missing VA fields to stay nil")
	}
	var historyCount int64
	if err = db.Model(&models.PaymentHistory{}).Where("payment_id = ?", payment.ID).Count(&historyCount).Error; err != nil {
		t.Fatalf("count history: %v", err)
	}
	if historyCount != 1 {
		t.Fatalf("expected one history entry, got %d", historyCount)
	}
	if len(producer.messages) != 1 {
		t.Fatalf("expected one kafka message, got %d", len(producer.messages))
	}
}

func TestWebhookHandlesSettlementWithVANumber(t *testing.T) {
	service, db, producer := newTestPaymentService(t)
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working dir: %v", err)
	}
	if err = os.Chdir("../.."); err != nil {
		t.Fatalf("change working dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDir); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
	originalGeneratePDF := generatePDFFromHTML
	generatePDFFromHTML = func(string, any) ([]byte, error) {
		return []byte("pdf"), nil
	}
	t.Cleanup(func() {
		generatePDFFromHTML = originalGeneratePDF
	})
	payment := createPayment(t, db, constants.Pending)
	webhook := signedWebhook(service, payment, constants.SettlementString)
	webhook.VANumbers = []dto.VANumber{{Bank: "bca", VANumber: "1234567890"}}

	err = service.Webhook(context.Background(), webhook)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	var stored models.Payment
	if err = db.First(&stored, payment.ID).Error; err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if *stored.Status != constants.Settlement {
		t.Fatalf("expected settlement status, got %v", *stored.Status)
	}
	if stored.PaidAt == nil {
		t.Fatalf("expected paid_at to be set")
	}
	if stored.VANumber == nil || *stored.VANumber != "1234567890" {
		t.Fatalf("expected va number to be stored, got %v", stored.VANumber)
	}
	if stored.Bank == nil || *stored.Bank != "bca" {
		t.Fatalf("expected bank to be stored, got %v", stored.Bank)
	}
	if stored.InvoiceLink == nil || *stored.InvoiceLink != "https://example.com/invoice.pdf" {
		t.Fatalf("expected invoice link to be stored, got %v", stored.InvoiceLink)
	}
	var historyCount int64
	if err = db.Model(&models.PaymentHistory{}).Where("payment_id = ?", payment.ID).Count(&historyCount).Error; err != nil {
		t.Fatalf("count history: %v", err)
	}
	if historyCount != 1 {
		t.Fatalf("expected one history entry, got %d", historyCount)
	}
	if len(producer.messages) != 1 {
		t.Fatalf("expected one kafka message, got %d", len(producer.messages))
	}
}

func TestWebhookHandlesExpire(t *testing.T) {
	service, db, producer := newTestPaymentService(t)
	payment := createPayment(t, db, constants.Pending)
	webhook := signedWebhook(service, payment, constants.ExpireString)

	err := service.Webhook(context.Background(), webhook)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	var stored models.Payment
	if err = db.First(&stored, payment.ID).Error; err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if *stored.Status != constants.Expire {
		t.Fatalf("expected expire status, got %v", *stored.Status)
	}
	if stored.PaidAt != nil {
		t.Fatalf("expected paid_at to remain nil")
	}
	if len(producer.messages) != 1 {
		t.Fatalf("expected one kafka message, got %d", len(producer.messages))
	}
}

func TestWebhookIgnoresDuplicateStatus(t *testing.T) {
	service, db, producer := newTestPaymentService(t)
	payment := createPayment(t, db, constants.Pending)
	webhook := signedWebhook(service, payment, constants.PendingString)

	err := service.Webhook(context.Background(), webhook)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	var historyCount int64
	if err = db.Model(&models.PaymentHistory{}).Where("payment_id = ?", payment.ID).Count(&historyCount).Error; err != nil {
		t.Fatalf("count history: %v", err)
	}
	if historyCount != 0 {
		t.Fatalf("expected no new history entry for duplicate status, got %d", historyCount)
	}
	if len(producer.messages) != 0 {
		t.Fatalf("expected no kafka message for duplicate status, got %d", len(producer.messages))
	}
}

func TestWebhookRejectsInvalidTransition(t *testing.T) {
	service, db, producer := newTestPaymentService(t)
	payment := createPayment(t, db, constants.Settlement)
	webhook := signedWebhook(service, payment, constants.PendingString)

	err := service.Webhook(context.Background(), webhook)

	if !errors.Is(err, errPayment.ErrWebhookInvalidTransition) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
	var stored models.Payment
	if err = db.First(&stored, payment.ID).Error; err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if *stored.Status != constants.Settlement {
		t.Fatalf("expected status to remain settlement, got %v", *stored.Status)
	}
	if len(producer.messages) != 0 {
		t.Fatalf("expected no kafka messages, got %d", len(producer.messages))
	}
}

var (
	_ gcs.IGSClient           = (*fakeGCSClient)(nil)
	_ kafka.IKafkaRegistry    = (*fakeKafkaRegistry)(nil)
	_ clients.IMidTransClient = (*fakeMidtransClient)(nil)
)
