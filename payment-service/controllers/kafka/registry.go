package kafka

type IKafkaRegistry interface {
	GetKafkaProducer() IKafka
}
type Registry struct {
	brokers []string
}

func (r Registry) GetKafkaProducer() IKafka {
	return NewKafkaProducer(r.brokers)
}

func NewKafkaRegistry(brokers []string) IKafkaRegistry {
	return &Registry{brokers: brokers}
}
