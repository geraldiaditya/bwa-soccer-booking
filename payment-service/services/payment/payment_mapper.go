package services

import (
	"payment-service/constants"
	"payment-service/domain/dto"
	"payment-service/domain/models"
)

func mapPaymentResponse(payment models.Payment) dto.PaymentResponse {
	status := constants.PaymentStatusString("")
	if payment.Status != nil {
		status = payment.Status.GetStatusString()
	}

	return dto.PaymentResponse{
		UUID:          payment.UUID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Status:        status,
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
	}
}

func mapPaymentResponses(payments []models.Payment) []dto.PaymentResponse {
	responses := make([]dto.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, mapPaymentResponse(payment))
	}
	return responses
}
