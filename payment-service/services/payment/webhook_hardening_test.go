package services

import (
	"context"
	"errors"
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

type hardeningKafkaRegistry struct {
	producer *hardeningKafkaProducer
}

func (h *hardeningKafkaRegistry) GetKafkaProducer() kafka.IKafka {
	return h.producer
}

func (h *hardeningKafkaRegistry) Close() error {
	return nil
}

type hardeningKafkaProducer struct {
	messages [][]byte
}

func (h *hardeningKafkaProducer) ProduceMessage(_ string, data []byte) error {
	h.messages = append(h.messages, data)
	return nil
}

func (h *hardeningKafkaProducer) Close() error {
	return nil
}

type hardeningGCSClient struct{}

func (h *hardeningGCSClient) UploadFile(context.Context, string, []byte) (string, error) {
	return "https://example.com/invoice.pdf", nil
}

type hardeningMidtransClient struct{}

func (h *hardeningMidtransClient) CreatePaymentLink(*dto.PaymentRequest) (*clients.MidTransData, error) {
	return &clients.MidTransData{RedirectURL: "https://example.com/pay"}, nil
}

func newWebhookHardeningPaymentService(t *testing.T) (*PaymentService, *gorm.DB, *hardeningKafkaProducer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&models.Payment{}, &models.PaymentHistory{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	producer := &hardeningKafkaProducer{}
	service := NewPaymentService(
		repositories.NewRepositoryRegistry(db),
		&hardeningGCSClient{},
		&hardeningKafkaRegistry{producer: producer},
		&hardeningMidtransClient{},
	).(*PaymentService)
	config2.Config.Midtrans.ServerKey = "server-key"
	config2.Config.Kafka.Topic = "payment-events"
	service.invoiceGenerator.renderPDF = func(*dto.InvoiceRequest) ([]byte, error) {
		return []byte("pdf"), nil
	}
	return service, db, producer
}

func createWebhookHardeningPayment(t *testing.T, db *gorm.DB, status constants.PaymentStatus) models.Payment {
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
	service, db, producer := newWebhookHardeningPaymentService(t)
	payment := createWebhookHardeningPayment(t, db, constants.Initial)
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
	service, _, _ := newWebhookHardeningPaymentService(t)
	tests := []struct {
		name    string
		webhook *dto.Webhook
	}{
		{name: "nil webhook", webhook: nil},
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
	service, db, producer := newWebhookHardeningPaymentService(t)
	payment := createWebhookHardeningPayment(t, db, constants.Initial)
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
	service, db, producer := newWebhookHardeningPaymentService(t)
	payment := createWebhookHardeningPayment(t, db, constants.Pending)
	webhook := signedWebhook(service, payment, constants.SettlementString)
	webhook.VANumbers = []dto.VANumber{{Bank: "bca", VANumber: "1234567890"}}

	err := service.Webhook(context.Background(), webhook)

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
	service, db, producer := newWebhookHardeningPaymentService(t)
	payment := createWebhookHardeningPayment(t, db, constants.Pending)
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
	service, db, producer := newWebhookHardeningPaymentService(t)
	payment := createWebhookHardeningPayment(t, db, constants.Pending)
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
	service, db, producer := newWebhookHardeningPaymentService(t)
	payment := createWebhookHardeningPayment(t, db, constants.Settlement)
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

func TestValidateStatusTransition(t *testing.T) {
	service, _, _ := newWebhookHardeningPaymentService(t)
	tests := []struct {
		name        string
		current     *constants.PaymentStatus
		next        constants.PaymentStatus
		expectedErr error
	}{
		{name: "initial to pending", current: statusPtr(constants.Initial), next: constants.Pending},
		{name: "initial to settlement", current: statusPtr(constants.Initial), next: constants.Settlement},
		{name: "initial to expire", current: statusPtr(constants.Initial), next: constants.Expire},
		{name: "pending to settlement", current: statusPtr(constants.Pending), next: constants.Settlement},
		{name: "pending to expire", current: statusPtr(constants.Pending), next: constants.Expire},
		{name: "duplicate pending", current: statusPtr(constants.Pending), next: constants.Pending},
		{name: "settlement to pending", current: statusPtr(constants.Settlement), next: constants.Pending, expectedErr: errPayment.ErrWebhookInvalidTransition},
		{name: "expire to pending", current: statusPtr(constants.Expire), next: constants.Pending, expectedErr: errPayment.ErrWebhookInvalidTransition},
		{name: "pending to initial", current: statusPtr(constants.Pending), next: constants.Initial, expectedErr: errPayment.ErrWebhookInvalidTransition},
		{name: "nil status", current: nil, next: constants.Pending, expectedErr: errPayment.ErrWebhookInvalidPayload},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &models.Payment{Status: tt.current}

			err := service.validateStatusTransition(payment, tt.next)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestValidateStatusTransitionRejectsNilPayment(t *testing.T) {
	service, _, _ := newWebhookHardeningPaymentService(t)

	err := service.validateStatusTransition(nil, constants.Pending)

	if !errors.Is(err, errPayment.ErrWebhookInvalidPayload) {
		t.Fatalf("expected invalid payload error, got %v", err)
	}
}

func statusPtr(status constants.PaymentStatus) *constants.PaymentStatus {
	return &status
}

var (
	_ gcs.IGSClient           = (*hardeningGCSClient)(nil)
	_ kafka.IKafkaRegistry    = (*hardeningKafkaRegistry)(nil)
	_ clients.IMidTransClient = (*hardeningMidtransClient)(nil)
)
