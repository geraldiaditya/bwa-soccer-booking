package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"order-service/common/utils"
	"order-service/domain/dto"
	"order-service/services"
)

const PaymentTopic = "payment-service-callback"

func NewPaymentKafka(service services.IServiceRegistry) IPaymentKafka {
	return &PaymentKafka{service: service}
}

type IPaymentKafka interface {
	HandlePayment(ctx context.Context, message *sarama.ConsumerMessage) error
}
type PaymentKafka struct {
	service services.IServiceRegistry
}

func (p *PaymentKafka) HandlePayment(ctx context.Context, message *sarama.ConsumerMessage) error {
	defer utils.Recover()
	var body dto.PaymentContent
	err := json.Unmarshal(message.Value, &body)
	if err != nil {
		logrus.Errorf("Error unmarshalling payment content %v", err)
		return err
	}
	data := body.Body.Data
	err = p.service.GetOrder().HandlePayment(ctx, &data)
	if err != nil {
		logrus.Errorf("Error handling payment %v", err)
		return err
	}
	logrus.Infof("HandlePayment Kafka success")
	return nil

}
