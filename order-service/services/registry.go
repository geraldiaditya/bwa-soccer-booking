package services

import (
	"order-service/clients"
	"order-service/repositories"
	services "order-service/services/order"
)

func NewServiceRegistry(repository repositories.IRepositoryRegistry, client clients.IClientRegistry) IServiceRegistry {
	return &Registry{
		repository: repository,
		client:     client,
	}
}

type IServiceRegistry interface {
	GetOrder() services.IOrderService
}

type Registry struct {
	repository repositories.IRepositoryRegistry
	client     clients.IClientRegistry
}

func (r *Registry) GetOrder() services.IOrderService {
	return services.NewOrderService(r.repository, r.client)
}
