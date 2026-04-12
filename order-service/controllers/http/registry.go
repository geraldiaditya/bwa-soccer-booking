package controllers

import (
	controllers "order-service/controllers/http/order"
	"order-service/services"
)

func NewControllerRegistry(service services.IServiceRegistry) IControllerRegistry {
	return &Registry{service: service}
}

type Registry struct {
	service services.IServiceRegistry
}

func (r *Registry) GetOrder() controllers.IOrderController {
	return controllers.NewOrderController(r.service)
}

type IControllerRegistry interface {
	GetOrder() controllers.IOrderController
}
