package clients

import (
	"order-service/clients/config"
	fieldClient "order-service/clients/field"
	paymentClient "order-service/clients/payment"
	userClient "order-service/clients/user"
	configApp "order-service/config"
)

type ClientRegistry struct{}

type IClientRegistry interface {
	GetUser() userClient.IUserClient
	GetPayment() paymentClient.IPaymentClient
	GetField() fieldClient.IFieldClient
}

func NewClientRegistry() IClientRegistry {
	return &ClientRegistry{}
}

func (registry *ClientRegistry) GetUser() userClient.IUserClient {
	return userClient.NewUserClient(
		config.NewClientConfig(
			config.WithBaseUrl(configApp.Config.InternalService.User.Host),
			config.WithSignatureKey(configApp.Config.InternalService.User.SignatureKey),
		))
}
func (registry *ClientRegistry) GetPayment() paymentClient.IPaymentClient {
	return paymentClient.NewPaymentClient(
		config.NewClientConfig(
			config.WithBaseUrl(configApp.Config.InternalService.User.Host),
			config.WithSignatureKey(configApp.Config.InternalService.User.SignatureKey),
		))
}

func (registry *ClientRegistry) GetField() fieldClient.IFieldClient {
	return fieldClient.NewFieldClient(
		config.NewClientConfig(
			config.WithBaseUrl(configApp.Config.InternalService.User.Host),
			config.WithSignatureKey(configApp.Config.InternalService.User.SignatureKey),
		))
}
