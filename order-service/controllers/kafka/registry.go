package kafka

import (
	kafka "order-service/controllers/kafka/payment"
	"order-service/services"
)

type Registry struct {
	service services.IServiceRegistry
}

func (k *Registry) GetPayment() kafka.IPaymentKafka {
	return kafka.NewPaymentKafka(k.service)
}

func NewKafkaRegistry(service services.IServiceRegistry) IKafkaRegistry {
	return &Registry{service: service}
}

type IKafkaRegistry interface {
	GetPayment() kafka.IPaymentKafka
}
