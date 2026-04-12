package clients

import (
	"payment-service/clients/config"
	clients "payment-service/clients/user"
	config2 "payment-service/config"
)

type ClientRegistry struct{}

type IClientRegistry interface {
	GetUser() clients.IUserClient
}

func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{}
}
func (registry *ClientRegistry) GetUser() clients.IUserClient {
	return clients.NewUserClient(
		config.NewClientConfig(
			config.WithBaseUrl(config2.Config.InternalService.User.Host),
			config.WithSignatureKey(config2.Config.InternalService.User.SignatureKey),
		))
}
