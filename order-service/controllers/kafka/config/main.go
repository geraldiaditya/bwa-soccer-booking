package kafka

import (
	"golang.org/x/exp/slices"
	"order-service/config"
	"order-service/controllers/kafka"
	kafkaPayment "order-service/controllers/kafka/payment"
)

func NewKafkaConsumer(consumer *ConsumerGroup, kafka kafka.IKafkaRegistry) IKafka {
	return &Kafka{
		consumer: consumer,
		kafka:    kafka,
	}
}

type IKafka interface {
	Register()
}

type Kafka struct {
	consumer *ConsumerGroup
	kafka    kafka.IKafkaRegistry
}

func (k *Kafka) Register() {
	k.paymentHandler()
}

func (k *Kafka) paymentHandler() {
	if slices.Contains(config.Config.Kafka.Topics, kafkaPayment.PaymentTopic) {
		k.consumer.RegisterHandler(kafkaPayment.PaymentTopic, k.kafka.GetPayment().HandlePayment)
	}
}
