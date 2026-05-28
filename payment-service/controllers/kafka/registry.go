package kafka

type IKafkaRegistry interface {
	GetKafkaProducer() IKafka
	Close() error
}
type Registry struct {
	producer IKafka
}

func (r Registry) GetKafkaProducer() IKafka {
	return r.producer
}

func (r Registry) Close() error {
	if r.producer == nil {
		return nil
	}
	return r.producer.Close()
}

func NewKafkaRegistry(brokers []string, maxRetry int) (IKafkaRegistry, error) {
	producer, err := NewKafkaProducer(brokers, maxRetry)
	if err != nil {
		return nil, err
	}
	return NewKafkaRegistryWithProducer(producer), nil
}

func NewKafkaRegistryWithProducer(producer IKafka) IKafkaRegistry {
	return &Registry{producer: producer}
}
