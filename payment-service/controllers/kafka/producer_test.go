package kafka

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

func TestProduceMessageSendsToKafka(t *testing.T) {
	config := sarama.NewConfig()
	producer := mocks.NewSyncProducer(t, config)
	producer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(func(message *sarama.ProducerMessage) error {
		if message.Topic != "payment-events" {
			t.Fatalf("expected topic payment-events, got %s", message.Topic)
		}
		value, err := message.Value.Encode()
		if err != nil {
			return err
		}
		if string(value) != "payload" {
			t.Fatalf("expected payload, got %s", string(value))
		}
		return nil
	})

	kafka := NewKafkaProducerWithSyncProducer(producer)

	if err := kafka.ProduceMessage("payment-events", []byte("payload")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := kafka.Close(); err != nil {
		t.Fatalf("expected close without error, got %v", err)
	}
}

func TestProduceMessageReturnsSendError(t *testing.T) {
	expectedErr := errors.New("broker unavailable")
	producer := mocks.NewSyncProducer(t, sarama.NewConfig())
	producer.ExpectSendMessageAndFail(expectedErr)

	kafka := NewKafkaProducerWithSyncProducer(producer)

	err := kafka.ProduceMessage("payment-events", []byte("payload"))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected send error %v, got %v", expectedErr, err)
	}
	if err := kafka.Close(); err != nil {
		t.Fatalf("expected close without error, got %v", err)
	}
}

func TestProduceMessageReturnsErrorWhenProducerIsNil(t *testing.T) {
	kafka := NewKafkaProducerWithSyncProducer(nil)

	err := kafka.ProduceMessage("payment-events", []byte("payload"))
	if err == nil {
		t.Fatal("expected error for nil producer, got nil")
	}
}
