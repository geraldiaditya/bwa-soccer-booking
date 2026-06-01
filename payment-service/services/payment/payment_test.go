package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	midtrans "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	config2 "payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
	paymentRepo "payment-service/repositories/payment"
	historyRepo "payment-service/repositories/payment_history"
)

type fakeRepositoryRegistry struct {
	payment paymentRepo.IPaymentRepository
	history historyRepo.IPaymentHistoryRepository
}

func (f *fakeRepositoryRegistry) GetPayment() paymentRepo.IPaymentRepository {
	return f.payment
}

func (f *fakeRepositoryRegistry) GetPaymentHistory() historyRepo.IPaymentHistoryRepository {
	return f.history
}

func (f *fakeRepositoryRegistry) GetTx() *gorm.DB {
	return nil
}

func (f *fakeRepositoryRegistry) WithTransaction(_ context.Context, fn func(repositories.IRepositoryRegistry) error) error {
	return fn(f)
}

type fakePaymentRepository struct {
	createFn        func(context.Context, *dto.PaymentRequest) (*models.Payment, error)
	findByOrderIDFn func(context.Context, string) (*models.Payment, error)
	updateFn        func(context.Context, string, *dto.UpdatePaymentRequest) (*models.Payment, error)

	createCalls int
	updateCalls int
}

func (f *fakePaymentRepository) FindAllWithPagination(context.Context, *dto.PaymentRequestParam) ([]models.Payment, int64, error) {
	return nil, 0, nil
}

func (f *fakePaymentRepository) FindByUUID(context.Context, string) (*models.Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	return f.findByOrderIDFn(ctx, orderID)
}

func (f *fakePaymentRepository) Create(ctx context.Context, request *dto.PaymentRequest) (*models.Payment, error) {
	f.createCalls++
	return f.createFn(ctx, request)
}

func (f *fakePaymentRepository) Update(ctx context.Context, orderID string, request *dto.UpdatePaymentRequest) (*models.Payment, error) {
	f.updateCalls++
	return f.updateFn(ctx, orderID, request)
}

type fakePaymentHistoryRepository struct {
	createFn    func(context.Context, *dto.PaymentHistoryRequest) error
	createCalls int
}

func (f *fakePaymentHistoryRepository) Create(ctx context.Context, request *dto.PaymentHistoryRequest) error {
	f.createCalls++
	return f.createFn(ctx, request)
}

type fakeMidtransClient struct {
	createFn    func(*dto.PaymentRequest) (*midtrans.MidTransData, error)
	createCalls int
}

func (f *fakeMidtransClient) CreatePaymentLink(request *dto.PaymentRequest) (*midtrans.MidTransData, error) {
	f.createCalls++
	return f.createFn(request)
}

type fakeUnexpectedGCSClient struct{}

func (f fakeUnexpectedGCSClient) UploadFile(context.Context, string, []byte) (string, error) {
	return "", errors.New("unexpected gcs call")
}

var _ gcs.IGSClient = fakeUnexpectedGCSClient{}

type fakePaymentServiceGCSClient struct {
	err      error
	calls    int
	filename string
	data     []byte
}

func (f *fakePaymentServiceGCSClient) UploadFile(_ context.Context, filename string, data []byte) (string, error) {
	f.calls++
	f.filename = filename
	f.data = data
	if f.err != nil {
		return "", f.err
	}
	return "https://storage.example/invoice.pdf", nil
}

type fakeKafkaRegistry struct {
	producer *fakeKafkaProducer
}

func (f *fakeKafkaRegistry) GetKafkaProducer() kafka.IKafka {
	return f.producer
}

func (f *fakeKafkaRegistry) Close() error {
	return nil
}

type fakeKafkaProducer struct {
	err      error
	calls    int
	topic    string
	messages [][]byte
}

func (f *fakeKafkaProducer) ProduceMessage(topic string, data []byte) error {
	f.calls++
	f.topic = topic
	f.messages = append(f.messages, data)
	return f.err
}

func (f *fakeKafkaProducer) Close() error {
	return nil
}

func TestPaymentServiceCreate(t *testing.T) {
	orderID := uuid.New()
	description := "field booking"
	paymentID := uuid.New()
	paymentStatus := constants.Initial
	createErr := errors.New("create payment failed")
	historyErr := errors.New("create history failed")
	midtransErr := errors.New("midtrans failed")

	tests := []struct {
		name              string
		expiredAt         time.Time
		midtransErr       error
		paymentCreateErr  error
		historyCreateErr  error
		wantErr           error
		wantMidtransCalls int
		wantCreateCalls   int
		wantHistoryCalls  int
	}{
		{
			name:              "success",
			expiredAt:         time.Now().Add(time.Hour),
			wantMidtransCalls: 1,
			wantCreateCalls:   1,
			wantHistoryCalls:  1,
		},
		{
			name:      "expired payment request",
			expiredAt: time.Now().Add(-time.Minute),
			wantErr:   errPayment.ErrExpireAtInvalid,
		},
		{
			name:              "midtrans failure",
			expiredAt:         time.Now().Add(time.Hour),
			midtransErr:       midtransErr,
			wantErr:           midtransErr,
			wantMidtransCalls: 1,
		},
		{
			name:              "payment create failure",
			expiredAt:         time.Now().Add(time.Hour),
			paymentCreateErr:  createErr,
			wantErr:           createErr,
			wantMidtransCalls: 1,
			wantCreateCalls:   1,
		},
		{
			name:              "history create failure",
			expiredAt:         time.Now().Add(time.Hour),
			historyCreateErr:  historyErr,
			wantErr:           historyErr,
			wantMidtransCalls: 1,
			wantCreateCalls:   1,
			wantHistoryCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepository := &fakePaymentRepository{}
			paymentRepository.createFn = func(_ context.Context, request *dto.PaymentRequest) (*models.Payment, error) {
				if request.PaymentLink != "https://pay.example/redirect" {
					t.Fatalf("expected midtrans redirect URL to be persisted, got %q", request.PaymentLink)
				}
				if tt.paymentCreateErr != nil {
					return nil, tt.paymentCreateErr
				}
				return &models.Payment{
					ID:          7,
					UUID:        paymentID,
					OrderID:     orderID,
					Amount:      request.Amount,
					Status:      &paymentStatus,
					PaymentLink: request.PaymentLink,
					Description: request.Description,
				}, nil
			}
			historyRepository := &fakePaymentHistoryRepository{
				createFn: func(_ context.Context, request *dto.PaymentHistoryRequest) error {
					if request.PaymentId != 7 || request.Status != constants.InitialString {
						t.Fatalf("unexpected history request: %+v", request)
					}
					return tt.historyCreateErr
				},
			}
			midtransClient := &fakeMidtransClient{
				createFn: func(*dto.PaymentRequest) (*midtrans.MidTransData, error) {
					if tt.midtransErr != nil {
						return nil, tt.midtransErr
					}
					return &midtrans.MidTransData{RedirectURL: "https://pay.example/redirect"}, nil
				},
			}

			service := NewPaymentService(
				&fakeRepositoryRegistry{payment: paymentRepository, history: historyRepository},
				fakeUnexpectedGCSClient{},
				&fakeKafkaRegistry{producer: &fakeKafkaProducer{}},
				midtransClient,
			)

			response, err := service.Create(context.Background(), &dto.PaymentRequest{
				OrderID:     orderID.String(),
				ExpiredAt:   tt.expiredAt,
				Amount:      125000,
				Description: &description,
			})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if response == nil || response.UUID != paymentID || response.PaymentLink != "https://pay.example/redirect" {
					t.Fatalf("unexpected response: %+v", response)
				}
			}
			if midtransClient.createCalls != tt.wantMidtransCalls {
				t.Fatalf("expected %d midtrans calls, got %d", tt.wantMidtransCalls, midtransClient.createCalls)
			}
			if paymentRepository.createCalls != tt.wantCreateCalls {
				t.Fatalf("expected %d payment create calls, got %d", tt.wantCreateCalls, paymentRepository.createCalls)
			}
			if historyRepository.createCalls != tt.wantHistoryCalls {
				t.Fatalf("expected %d history create calls, got %d", tt.wantHistoryCalls, historyRepository.createCalls)
			}
		})
	}
}

func TestPaymentServiceWebhook(t *testing.T) {
	orderID := uuid.New()
	paymentID := uuid.New()
	description := "field booking"
	expiredAt := time.Now().Add(24 * time.Hour)
	notFoundErr := errPayment.ErrPaymentNotFound
	invoiceErr := errors.New("invoice render failed")
	uploadErr := errors.New("invoice upload failed")
	kafkaErr := errors.New("kafka publish failed")

	tests := []struct {
		name             string
		status           constants.PaymentStatusString
		findErr          error
		invoiceErr       error
		uploadErr        error
		kafkaErr         error
		wantErr          error
		wantUpdateCalls  int
		wantHistoryCalls int
		wantInvoiceCalls int
		wantUploadCalls  int
		wantKafkaCalls   int
		wantKafkaStatus  string
		wantKafkaPaidAt  bool
	}{
		{
			name:             "settlement success",
			status:           constants.SettlementString,
			wantUpdateCalls:  2,
			wantHistoryCalls: 1,
			wantInvoiceCalls: 1,
			wantUploadCalls:  1,
			wantKafkaCalls:   1,
			wantKafkaStatus:  "settlement",
			wantKafkaPaidAt:  true,
		},
		{
			name:             "pending status update",
			status:           constants.PendingString,
			wantUpdateCalls:  1,
			wantHistoryCalls: 1,
			wantKafkaCalls:   1,
			wantKafkaStatus:  "pending",
		},
		{
			name:             "expire status update",
			status:           constants.ExpireString,
			wantUpdateCalls:  1,
			wantHistoryCalls: 1,
			wantKafkaCalls:   1,
			wantKafkaStatus:  "expire",
		},
		{
			name:    "payment not found",
			status:  constants.SettlementString,
			findErr: notFoundErr,
			wantErr: notFoundErr,
		},
		{
			name:             "invoice generation failure",
			status:           constants.SettlementString,
			invoiceErr:       invoiceErr,
			wantErr:          invoiceErr,
			wantUpdateCalls:  1,
			wantHistoryCalls: 1,
			wantInvoiceCalls: 1,
		},
		{
			name:             "invoice upload failure",
			status:           constants.SettlementString,
			uploadErr:        uploadErr,
			wantErr:          uploadErr,
			wantUpdateCalls:  1,
			wantHistoryCalls: 1,
			wantInvoiceCalls: 1,
			wantUploadCalls:  1,
		},
		{
			name:             "kafka publish failure",
			status:           constants.SettlementString,
			kafkaErr:         kafkaErr,
			wantErr:          kafkaErr,
			wantUpdateCalls:  2,
			wantHistoryCalls: 1,
			wantInvoiceCalls: 1,
			wantUploadCalls:  1,
			wantKafkaCalls:   1,
			wantKafkaStatus:  "settlement",
			wantKafkaPaidAt:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentStatus := constants.Initial
			currentPayment := &models.Payment{
				ID:          11,
				UUID:        paymentID,
				OrderID:     orderID,
				Amount:      275000,
				Status:      &currentStatus,
				Description: &description,
				ExpiredAt:   &expiredAt,
			}

			paymentRepository := &fakePaymentRepository{}
			paymentRepository.findByOrderIDFn = func(_ context.Context, id string) (*models.Payment, error) {
				if id != orderID.String() {
					t.Fatalf("unexpected order ID lookup %q", id)
				}
				if tt.findErr != nil {
					return nil, tt.findErr
				}
				return currentPayment, nil
			}
			paymentRepository.updateFn = func(_ context.Context, id string, request *dto.UpdatePaymentRequest) (*models.Payment, error) {
				if id != orderID.String() {
					t.Fatalf("unexpected order ID update %q", id)
				}
				if request.Status != nil {
					currentPayment.Status = request.Status
				}
				if request.TransactionId != nil {
					currentPayment.TransactionID = request.TransactionId
				}
				if request.PaidAt != nil {
					currentPayment.PaidAt = request.PaidAt
				}
				if request.VANumber != nil {
					currentPayment.VANumber = request.VANumber
				}
				if request.Bank != nil {
					currentPayment.Bank = request.Bank
				}
				if request.InvoiceLink != nil {
					currentPayment.InvoiceLink = request.InvoiceLink
				}
				currentPayment.Acquirer = &request.Acquirer
				return currentPayment, nil
			}
			historyRepository := &fakePaymentHistoryRepository{
				createFn: func(_ context.Context, request *dto.PaymentHistoryRequest) error {
					if request.PaymentId != currentPayment.ID {
						t.Fatalf("unexpected history payment ID %d", request.PaymentId)
					}
					return nil
				},
			}
			producer := &fakeKafkaProducer{err: tt.kafkaErr}
			gcsClient := &fakePaymentServiceGCSClient{err: tt.uploadErr}
			service := NewPaymentService(
				&fakeRepositoryRegistry{payment: paymentRepository, history: historyRepository},
				gcsClient,
				&fakeKafkaRegistry{producer: producer},
				&fakeMidtransClient{},
			).(*PaymentService)
			invoiceCalls := 0
			service.invoiceGenerator.renderPDF = func(request *dto.InvoiceRequest) ([]byte, error) {
				invoiceCalls++
				if request.Data.PaymentDetail.BankName != "BCA" || request.Data.PaymentDetail.VANumber != "123456" {
					t.Fatalf("unexpected invoice request: %+v", request)
				}
				if tt.invoiceErr != nil {
					return nil, tt.invoiceErr
				}
				return []byte("pdf"), nil
			}

			config2.Config.Midtrans.ServerKey = "server-key"
			webhook := &dto.Webhook{
				OrderId:           orderID,
				TransactionStatus: tt.status,
				TransactionId:     "trx-123",
				StatusCode:        "200",
				GrossAmount:       "275000.00",
				VANumbers: []dto.VANumber{{
					VANumber: "123456",
					Bank:     "bca",
				}},
				PaymentType: "bank_transfer",
				Acquirer:    "bca",
			}
			webhook.SignatureKey = service.webhookSignature(
				webhook.OrderId.String(),
				webhook.StatusCode,
				webhook.GrossAmount,
				config2.Config.Midtrans.ServerKey,
			)

			err := service.Webhook(context.Background(), webhook)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if paymentRepository.updateCalls != tt.wantUpdateCalls {
				t.Fatalf("expected %d update calls, got %d", tt.wantUpdateCalls, paymentRepository.updateCalls)
			}
			if historyRepository.createCalls != tt.wantHistoryCalls {
				t.Fatalf("expected %d history calls, got %d", tt.wantHistoryCalls, historyRepository.createCalls)
			}
			if invoiceCalls != tt.wantInvoiceCalls {
				t.Fatalf("expected %d invoice calls, got %d", tt.wantInvoiceCalls, invoiceCalls)
			}
			if gcsClient.calls != tt.wantUploadCalls {
				t.Fatalf("expected %d upload calls, got %d", tt.wantUploadCalls, gcsClient.calls)
			}
			if tt.wantUploadCalls > 0 && (gcsClient.filename == "" || string(gcsClient.data) != "pdf") {
				t.Fatalf("unexpected upload args filename=%q pdf=%q", gcsClient.filename, string(gcsClient.data))
			}
			if producer.calls != tt.wantKafkaCalls {
				t.Fatalf("expected %d kafka calls, got %d", tt.wantKafkaCalls, producer.calls)
			}
			if tt.wantKafkaCalls > 0 {
				var message dto.KafkaMessage
				if err := json.Unmarshal(producer.messages[0], &message); err != nil {
					t.Fatalf("invalid kafka payload: %v", err)
				}
				if message.Body.Data.Status != tt.wantKafkaStatus {
					t.Fatalf("expected kafka status %q, got %q", tt.wantKafkaStatus, message.Body.Data.Status)
				}
				if (message.Body.Data.PaidAt != nil) != tt.wantKafkaPaidAt {
					t.Fatalf("expected kafka paidAt present=%v, got %+v", tt.wantKafkaPaidAt, message.Body.Data.PaidAt)
				}
			}
		})
	}
}
