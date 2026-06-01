package services

import (
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/json"
	"fmt"
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

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
		repository:       repository,
		kafka:            kafka,
		midtrans:         midtrans,
		invoiceGenerator: NewInvoiceGenerator(gcs),
	}
}

type PaymentService struct {
	repository       repositories.IRepositoryRegistry
	kafka            kafka.IKafkaRegistry
	midtrans         clients.IMidTransClient
	invoiceGenerator *InvoiceGenerator
}

var marshalKafkaMessage = func(message dto.KafkaMessage) ([]byte, error) {
	return json.Marshal(message)
}

func (p *PaymentService) GetAllWithPagination(ctx context.Context, param *dto.PaymentRequestParam) (*utils.PaginationResult, error) {
	payments, total, err := p.repository.GetPayment().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}
	paymentResults := mapPaymentResponses(payments)

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
	response := mapPaymentResponse(payment)
	return &response, nil
}

func (p *PaymentService) Create(ctx context.Context, request *dto.PaymentRequest) (*dto.PaymentResponse, error) {
	var (
		txErr, err error
		payment    *models.Payment
		response   *dto.PaymentResponse
		midtrans   *clients.MidTransData
	)

	err = p.repository.WithTransaction(ctx, func(repository repositories.IRepositoryRegistry) error {
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
		payment, txErr = repository.GetPayment().Create(ctx, paymentRequest)
		if txErr != nil {
			return txErr
		}

		txErr = repository.GetPaymentHistory().Create(ctx, &dto.PaymentHistoryRequest{
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

	mappedResponse := mapCreatedPaymentResponse(payment)
	response = &mappedResponse
	return response, nil
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

func (p *PaymentService) validateWebhookPayload(webhook *dto.Webhook) error {
	if webhook == nil {
		return errPayment.ErrWebhookInvalidPayload
	}
	if webhook.OrderId == uuid.Nil ||
		webhook.TransactionId == "" ||
		webhook.StatusCode == "" ||
		webhook.GrossAmount == "" ||
		webhook.SignatureKey == "" {
		return errPayment.ErrWebhookInvalidPayload
	}
	if !p.isSupportedTransactionStatus(webhook.TransactionStatus) {
		return errPayment.ErrWebhookInvalidPayload
	}
	return nil
}

func (p *PaymentService) isSupportedTransactionStatus(status constants.PaymentStatusString) bool {
	switch status {
	case constants.PendingString, constants.SettlementString, constants.ExpireString:
		return true
	default:
		return false
	}
}

func (p *PaymentService) verifyWebhookSignature(webhook *dto.Webhook) error {
	serverKey := config2.Config.Midtrans.ServerKey
	if serverKey == "" {
		return errPayment.ErrWebhookInvalidSignature
	}
	expected := p.webhookSignature(webhook.OrderId.String(), webhook.StatusCode, webhook.GrossAmount, serverKey)
	if subtle.ConstantTimeCompare([]byte(webhook.SignatureKey), []byte(expected)) != 1 {
		return errPayment.ErrWebhookInvalidSignature
	}
	return nil
}

func (p *PaymentService) webhookSignature(orderID, statusCode, grossAmount, serverKey string) string {
	signaturePayload := orderID + statusCode + grossAmount + serverKey
	signature := sha512.Sum512([]byte(signaturePayload))
	return fmt.Sprintf("%x", signature)
}

func (p *PaymentService) isDuplicateStatus(payment *models.Payment, next constants.PaymentStatus) bool {
	return payment != nil && payment.Status != nil && *payment.Status == next
}

func (p *PaymentService) validateStatusTransition(payment *models.Payment, next constants.PaymentStatus) error {
	if payment == nil || payment.Status == nil {
		return errPayment.ErrWebhookInvalidPayload
	}
	if *payment.Status == next {
		return nil
	}
	switch *payment.Status {
	case constants.Initial:
		if next == constants.Pending || next == constants.Settlement || next == constants.Expire {
			return nil
		}
	case constants.Pending:
		if next == constants.Settlement || next == constants.Expire {
			return nil
		}
	case constants.Settlement, constants.Expire:
		return errPayment.ErrWebhookInvalidTransition
	}
	return errPayment.ErrWebhookInvalidTransition
}

func (p *PaymentService) webhookVirtualAccount(webhook *dto.Webhook) (*string, *string) {
	if webhook == nil || len(webhook.VANumbers) == 0 {
		return nil, nil
	}
	vaNumber := webhook.VANumbers[0].VANumber
	bank := webhook.VANumbers[0].Bank
	if vaNumber == "" && bank == "" {
		return nil, nil
	}
	return &vaNumber, &bank
}

func (p *PaymentService) produceToKafka(
	request *dto.Webhook,
	payment *models.Payment,
	paidAt *time.Time) error {
	var expiredAt time.Time
	if payment.ExpiredAt != nil {
		expiredAt = *payment.ExpiredAt
	}
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
			ExpiredAt: expiredAt,
		},
	}
	kafkaMessage := dto.KafkaMessage{
		Event:    event,
		Body:     body,
		MetaData: metadata,
	}
	topic := config2.Config.Kafka.Topic
	kafkaMessageJson, err := marshalKafkaMessage(kafkaMessage)
	if err != nil {
		logrus.WithError(err).Error("failed to serialize payment kafka event")
		return fmt.Errorf("serialize payment kafka event: %w", err)
	}
	if err := p.kafka.GetKafkaProducer().ProduceMessage(topic, kafkaMessageJson); err != nil {
		logrus.WithError(err).WithField("topic", topic).Error("failed to publish payment kafka event")
		return fmt.Errorf("publish payment kafka event: %w", err)
	}
	return nil
}

func (p *PaymentService) Webhook(ctx context.Context, webhook *dto.Webhook) error {
	var (
		txErr, err          error
		paymentBeforeUpdate *models.Payment
		paymentAfterUpdate  *models.Payment
		paidAt              *time.Time
		invoiceLink         string
		updated             bool
	)

	if err = p.validateWebhookPayload(webhook); err != nil {
		return err
	}
	if err = p.verifyWebhookSignature(webhook); err != nil {
		return err
	}

	err = p.repository.WithTransaction(ctx, func(repository repositories.IRepositoryRegistry) error {
		paymentBeforeUpdate, txErr = repository.GetPayment().FindByOrderID(ctx, webhook.OrderId.String())
		if txErr != nil {
			return txErr
		}

		status := webhook.TransactionStatus.GetStatusInt()
		if txErr = p.validateStatusTransition(paymentBeforeUpdate, status); txErr != nil {
			return txErr
		}
		if p.isDuplicateStatus(paymentBeforeUpdate, status) {
			paymentAfterUpdate = paymentBeforeUpdate
			return nil
		}

		if webhook.TransactionStatus == constants.SettlementString {
			now := time.Now()
			paidAt = &now
		}
		vaNumber, bank := p.webhookVirtualAccount(webhook)
		_, txErr = repository.GetPayment().Update(ctx, webhook.OrderId.String(), &dto.UpdatePaymentRequest{
			TransactionId: &webhook.TransactionId,
			Status:        &status,
			PaidAt:        paidAt,
			VANumber:      vaNumber,
			Bank:          bank,
			Acquirer:      webhook.Acquirer,
		})
		if txErr != nil {
			return txErr
		}
		updated = true
		afterUpdate := *paymentBeforeUpdate
		paymentAfterUpdate = &afterUpdate
		paymentAfterUpdate.TransactionID = &webhook.TransactionId
		paymentAfterUpdate.Status = &status
		paymentAfterUpdate.PaidAt = paidAt
		paymentAfterUpdate.VANumber = vaNumber
		paymentAfterUpdate.Bank = bank
		paymentAfterUpdate.Acquirer = &webhook.Acquirer

		txErr = repository.GetPaymentHistory().Create(ctx, &dto.PaymentHistoryRequest{
			PaymentId: paymentAfterUpdate.ID,
			Status:    paymentAfterUpdate.Status.GetStatusString(),
		})
		if txErr != nil {
			return txErr
		}

		if webhook.TransactionStatus == constants.SettlementString {
			invoiceNumber := BuildInvoiceNumber(time.Now(), paymentAfterUpdate.OrderID, paymentAfterUpdate.UUID)
			invoiceRequest := BuildInvoiceRequest(paymentAfterUpdate, webhook, *paidAt, invoiceNumber)
			invoiceLink, txErr = p.invoiceGenerator.GenerateAndUpload(ctx, invoiceRequest)
			if txErr != nil {
				return txErr
			}
			_, txErr = repository.GetPayment().Update(ctx, webhook.OrderId.String(), &dto.UpdatePaymentRequest{
				InvoiceLink: &invoiceLink,
			})
			if txErr != nil {
				return txErr
			}
			paymentAfterUpdate.InvoiceLink = &invoiceLink
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}
	// Kafka publish happens after the database transaction commits. If this
	// fails, payment state remains updated and callers receive an explicit error
	// so the webhook can be retried by the sender.
	err = p.produceToKafka(webhook, paymentAfterUpdate, paidAt)
	if err != nil {
		return err
	}
	return nil
}
