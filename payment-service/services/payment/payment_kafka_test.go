package services

import (
	"errors"
	"testing"
	"time"

	configApp "payment-service/config"
	"payment-service/constants"
	"payment-service/controllers/kafka"
	"payment-service/domain/dto"
	"payment-service/domain/models"

	"github.com/google/uuid"
)

type stubKafkaRegistry struct {
	producer *stubKafkaProducer
}

func (s stubKafkaRegistry) GetKafkaProducer() kafka.IKafka {
	return s.producer
}

func (s stubKafkaRegistry) Close() error {
	return s.producer.Close()
}

type stubKafkaProducer struct {
	topic      string
	data       []byte
	publishErr error
	publishes  int
	closeCalls int
}

func (s *stubKafkaProducer) ProduceMessage(topic string, data []byte) error {
	s.publishes++
	s.topic = topic
	s.data = data
	return s.publishErr
}

func (s *stubKafkaProducer) Close() error {
	s.closeCalls++
	return nil
}

func TestProduceToKafkaPublishesPaymentEvent(t *testing.T) {
	configApp.Config.Kafka.Topic = "payment-events"
	expiredAt := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	paidAt := time.Date(2026, 5, 28, 10, 0, 0, 0, time.UTC)
	producer := &stubKafkaProducer{}
	service := &PaymentService{kafka: stubKafkaRegistry{producer: producer}}
	orderID := uuid.New()
	paymentID := uuid.New()

	err := service.produceToKafka(
		&dto.Webhook{
			OrderId:           orderID,
			TransactionStatus: constants.SettlementString,
		},
		&models.Payment{
			UUID:      paymentID,
			OrderID:   orderID,
			ExpiredAt: &expiredAt,
		},
		&paidAt,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if producer.topic != "payment-events" {
		t.Fatalf("expected topic payment-events, got %s", producer.topic)
	}
	if len(producer.data) == 0 {
		t.Fatal("expected serialized kafka payload")
	}
}

func TestProduceToKafkaReturnsPublishError(t *testing.T) {
	configApp.Config.Kafka.Topic = "payment-events"
	expectedErr := errors.New("publish failed")
	expiredAt := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	producer := &stubKafkaProducer{publishErr: expectedErr}
	service := &PaymentService{kafka: stubKafkaRegistry{producer: producer}}
	orderID := uuid.New()

	err := service.produceToKafka(
		&dto.Webhook{
			OrderId:           orderID,
			TransactionStatus: constants.ExpireString,
		},
		&models.Payment{
			UUID:      uuid.New(),
			OrderID:   orderID,
			ExpiredAt: &expiredAt,
		},
		nil,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected publish error %v, got %v", expectedErr, err)
	}
}

func TestProduceToKafkaReturnsSerializationError(t *testing.T) {
	expectedErr := errors.New("serialize failed")
	originalMarshalKafkaMessage := marshalKafkaMessage
	marshalKafkaMessage = func(dto.KafkaMessage) ([]byte, error) {
		return nil, expectedErr
	}
	t.Cleanup(func() {
		marshalKafkaMessage = originalMarshalKafkaMessage
	})

	configApp.Config.Kafka.Topic = "payment-events"
	expiredAt := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	producer := &stubKafkaProducer{}
	service := &PaymentService{kafka: stubKafkaRegistry{producer: producer}}
	orderID := uuid.New()

	err := service.produceToKafka(
		&dto.Webhook{
			OrderId:           orderID,
			TransactionStatus: constants.SettlementString,
		},
		&models.Payment{
			UUID:      uuid.New(),
			OrderID:   orderID,
			ExpiredAt: &expiredAt,
		},
		nil,
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected serialization error %v, got %v", expectedErr, err)
	}
	if producer.publishes != 0 {
		t.Fatalf("expected producer not to be called, got %d publishes", producer.publishes)
	}
}
