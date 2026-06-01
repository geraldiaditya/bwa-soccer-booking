package error

import "errors"

var (
	ErrPaymentNotFound          = errors.New("payment not found")
	ErrExpireAtInvalid          = errors.New("expired time must be greater than current time")
	ErrWebhookInvalidPayload    = errors.New("invalid payment webhook payload")
	ErrWebhookInvalidSignature  = errors.New("invalid payment webhook signature")
	ErrWebhookInvalidTransition = errors.New("invalid payment status transition")
)

var PaymentErrors = []error{
	ErrPaymentNotFound,
	ErrExpireAtInvalid,
	ErrWebhookInvalidPayload,
	ErrWebhookInvalidSignature,
	ErrWebhookInvalidTransition,
}
