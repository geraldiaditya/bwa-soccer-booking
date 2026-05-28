package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type Kafka struct {
	producer sarama.SyncProducer
}

type IKafka interface {
	ProduceMessage(string, []byte) error
	Close() error
}

func NewKafkaProducer(brokers []string, maxRetry int) (IKafka, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	config.Producer.Retry.Max = maxRetry

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return NewKafkaProducerWithSyncProducer(producer), nil
}

func NewKafkaProducerWithSyncProducer(producer sarama.SyncProducer) IKafka {
	return &Kafka{producer: producer}
}

func (k *Kafka) ProduceMessage(topic string, data []byte) error {
	if k == nil || k.producer == nil {
		return fmt.Errorf("kafka producer is not initialized")
	}

	message := &sarama.ProducerMessage{
		Topic:   topic,
		Headers: nil,
		Value:   sarama.ByteEncoder(data),
	}

	partition, offset, err := k.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("send kafka message to topic %q: %w", topic, err)
	}
	logrus.Infof("Message sent to topic(%s)/partition(%d) at offset (%d)", topic, partition, offset)
	return nil
}

func (k *Kafka) Close() error {
	if k == nil || k.producer == nil {
		return nil
	}

	if err := k.producer.Close(); err != nil {
		return fmt.Errorf("close kafka producer: %w", err)
	}
	return nil
}
