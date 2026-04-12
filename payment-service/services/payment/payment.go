package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/rand"
	"os"
	clients "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	"payment-service/common/utils"
	config2 "payment-service/config"
	"payment-service/constants"
	errPayment "payment-service/constants/error/payment"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"
	"payment-service/repositories"
	"strings"
	"time"
)

type IPaymentService interface {
	GetAllWithPagination(context.Context, *dto.PaymentRequestParam) (*utils.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.PaymentResponse, error)
	Create(context.Context, *dto.PaymentRequest) (*dto.PaymentResponse, error)
	Webhook(context.Context, *dto.Webhook) error
}

func NewPaymentService(
	repository repositories.IRepositoryRegistry,
	gcs gcs.IGSClient,
	kafka kafka.IKafkaRegistry,
	midtrans clients.IMidTransClient,
) IPaymentService {
	return &PaymentService{
		repository: repository,
		gcs:        gcs,
		kafka:      kafka,
		midtrans:   midtrans,
	}
}

type PaymentService struct {
	repository repositories.IRepositoryRegistry
	gcs        gcs.IGSClient
	kafka      kafka.IKafkaRegistry
	midtrans   clients.IMidTransClient
}

func (p *PaymentService) GetAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) (*utils.PaginationResult, error) {
	payments, total, err := p.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}
	paymentResults := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResults = append(paymentResults, dto.PaymentResponse{
			UUID:          payment.UUID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			Status:        payment.Status.GetStatusString(),
			PaymentLink:   payment.PaymentLink,
			TransactionId: payment.TransactionID,
			PaidAt:        payment.PaidAt,
			VANumber:      payment.VANumber,
			Bank:          payment.Bank,
			InvoiceLink:   payment.InvoiceLink,
			Acquirer:      payment.Acquirer,
			Description:   payment.Description,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	paginationParam := utils.PaginationParam{
		Count: total,
		Page:  param.Page,
		Limit: param.Limit,
		Data:  paymentResults,
	}
	response := utils.GeneratePagination(paginationParam)
	return &response, nil
}

func (p *PaymentService) GetByUUID(ctx context.Context, s string) (*dto.PaymentResponse, error) {
	payment, err := p.repository.GetPayment().FindByUUID(ctx, s)
	if err != nil {
		return nil, err
	}
	return &dto.PaymentResponse{
		UUID:          payment.UUID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        payment.Status.GetStatusString(),
		PaymentLink:   payment.PaymentLink,
		TransactionId: payment.TransactionID,
		PaidAt:        payment.PaidAt,
		VANumber:      payment.VANumber,
		Bank:          payment.Bank,
		InvoiceLink:   payment.InvoiceLink,
		Acquirer:      payment.Acquirer,
		Description:   payment.Description,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}

func (p *PaymentService) Create(ctx context.Context, request *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment
		response   *dto.PaymentResponse
		midtrans   *clients.MidTransData
	)

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		if !request.ExpiredAt.After(time.Now()) {
			return errPayment.ErrExpireAtInvalid
		}
		midtrans, txErr = p.midtrans.CreatePaymentLink(request)
		if txErr != nil {
			return txErr
		}
		paymentRequest := &dto.PaymentRequest{
			OrderID:     request.OrderID,
			Amount:      request.Amount,
			Description: request.Description,
			ExpiredAt:   request.ExpiredAt,
			PaymentLink: midtrans.RedirectURL,
		}
		payment, txErr = p.repository.GetPayment().Create(ctx, tx, paymentRequest)
		if txErr != nil {
			return txErr
		}

		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentId: payment.ID,
			Status:    payment.Status.GetStatusString(),
		})
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	response = &dto.PaymentResponse{
		UUID:        payment.UUID,
		OrderID:     payment.OrderID,
		Amount:      payment.Amount,
		Status:      payment.Status.GetStatusString(),
		PaymentLink: payment.PaymentLink,
		Description: payment.Description,
	}
	return response, nil
}

func (p *PaymentService) convertToIndonesiaMonth(englishMonth string) string {
	monthMap := map[string]string{
		"January":  "Januari",
		"February": "Februari",
		"March":    "Maret",
		"April":    "April",

		"May":       "May",
		"June":      "Juni",
		"July":      "Juli",
		"August":    "Agustus",
		"September": "September",
		"October":   "OKtober",
		"November":  "November",
		"December":  "Desember",
	}

	indonesiaMonth, ok := monthMap[englishMonth]
	if !ok {
		return errors.New("month not found").Error()
	}
	return indonesiaMonth
}

func (p *PaymentService) generatePDF(req *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplatePath := "templates/invoice.html"
	htmlTemplate, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	jsonData, _ := json.Marshal(req)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	pdf, err := utils.GeneratePDFFromHTML(string(htmlTemplate), data)
	if err != nil {
		return nil, err
	}
	return pdf, nil
}

func (p *PaymentService) uploadToGCS(ctx context.Context, invoiceNumber string, pdf []byte) (string, error) {
	invoiceNumberReplace := strings.ToLower(strings.ReplaceAll(invoiceNumber, "/", "-"))
	filename := fmt.Sprintf("%s.pdf", invoiceNumberReplace)
	url, err := p.gcs.UploadFile(ctx, filename, pdf)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (p *PaymentService) randomNumber() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	number := random.Intn(900000) + 100000
	return number
}

func (p *PaymentService) mapTransactionStatusToEvent(status constants.PaymentStatusString) string {
	var paymentStatus string
	switch status {
	case constants.PendingString:
		paymentStatus = strings.ToUpper(constants.PendingString.String())
	case constants.SettlementString:
		paymentStatus = strings.ToUpper(constants.SettlementString.String())
	case constants.ExpireString:
		paymentStatus = strings.ToUpper(constants.ExpireString.String())
	}
	return paymentStatus
}

func (p *PaymentService) produceToKafka(
	request *dto.Webhook,
	payment *models.Payment,
	paidAt *time.Time) error {
	event := dto.KafkaEvent{
		Name: p.mapTransactionStatusToEvent(request.TransactionStatus),
	}
	metadata := dto.KafkaMetaData{
		Sender:    "payment-service",
		SendingAt: time.Now().Format(time.RFC3339),
	}
	body := dto.KafkaBody{
		Type: "JSON",
		Data: &dto.KafkaData{
			OrderID:   payment.OrderID,
			PaymentID: payment.UUID,
			Status:    request.TransactionStatus.String(),
			PaidAt:    paidAt,
			ExpiredAt: *payment.ExpiredAt,
		},
	}
	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		Body:     body,
		MetaData: metadata,
	}
	topic := config2.Config.Kafka.Topic
	kafkaMessageJson, _ := json.Marshal(kafkaMessage)
	err := p.kafka.GetKafkaProducer().ProduceMessage(topic, kafkaMessageJson)
	if err != nil {
		return err
	}
	return nil
}

func (p *PaymentService) Webhook(ctx context.Context, webhook *dto.Webhook) error {
	var (
		txErr, err         error
		paymentAfterUpdate *models.Payment
		paidAt             *time.Time
		invoiceLink        string
		pdf                []byte
	)

	err = p.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		_, txErr = p.repository.GetPayment().FindByOrderID(ctx, webhook.OrderId.String())
		if txErr != nil {
			return txErr
		}

		if webhook.TransactionStatus == constants.SettlementString {
			now := time.Now()
			paidAt = &now
		}
		status := webhook.TransactionStatus.GetStatusInt()
		vaNumber := webhook.VANumbers[0].VANumber
		bank := webhook.VANumbers[0].Bank
		_, txErr = p.repository.GetPayment().Update(ctx, tx, webhook.OrderId.String(), &dto.UpdatePaymentRequest{
			TransactionId: &webhook.TransactionId,
			Status:        &status,
			PaidAt:        paidAt,
			VANumber:      &vaNumber,
			Bank:          &bank,
			Acquirer:      webhook.Acquirer,
		})
		if txErr != nil {
			return txErr
		}
		paymentAfterUpdate, txErr = p.repository.GetPayment().FindByOrderID(ctx, webhook.OrderId.String())
		if txErr != nil {
			return txErr
		}
		txErr = p.repository.GetPaymentHistory().Create(ctx, tx, &dto.PaymentHistoryRequest{
			PaymentId: paymentAfterUpdate.ID,
			Status:    paymentAfterUpdate.Status.GetStatusString(),
		})

		if webhook.TransactionStatus == constants.SettlementString {
			paidDay := paidAt.Format("02")
			paidMonth := paidAt.Format("January")
			paidYear := paidAt.Format("2006")
			invoiceNumber := fmt.Sprintf("INV/%s/ORD/%d", time.Now().Format(time.DateOnly), p.randomNumber())
			total := utils.RupiahFormat(&paymentAfterUpdate.Amount)
			invoiceRequest := &dto.InvoiceRequest{
				InvoiceNumber: invoiceNumber,
				Data: dto.InvoiceData{
					PaymentDetail: dto.InvoicePaymentDetail{
						BankName:      strings.ToUpper(*paymentAfterUpdate.Bank),
						PaymentMethod: webhook.PaymentType,
						VANumber:      *paymentAfterUpdate.VANumber,
						Date:          fmt.Sprintf("%s %s %s", paidDay, paidMonth, paidYear),
						IsPaid:        true,
					},
					Items: []dto.InvoiceItem{
						{
							Description: *paymentAfterUpdate.Description,
							Price:       total,
						},
					},
					Total: total,
				},
			}
			pdf, txErr = p.generatePDF(invoiceRequest)
			if txErr != nil {
				return txErr
			}
			invoiceLink, txErr = p.uploadToGCS(ctx, invoiceNumber, pdf)
			if txErr != nil {
				return txErr
			}
			_, txErr = p.repository.GetPayment().Update(ctx, tx, webhook.OrderId.String(), &dto.UpdatePaymentRequest{
				InvoiceLink: &invoiceLink,
			})
			if txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = p.produceToKafka(webhook, paymentAfterUpdate, paidAt)
	if err != nil {
		return err
	}
	return nil
}
