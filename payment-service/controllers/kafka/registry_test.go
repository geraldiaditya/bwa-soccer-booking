package kafka

import (
	"errors"
	"testing"
)

type fakeProducer struct {
	closeCalls int
}

func (f *fakeProducer) ProduceMessage(string, []byte) error {
	return nil
}

func (f *fakeProducer) Close() error {
	f.closeCalls++
	return nil
}

func TestNewKafkaRegistryWithProducerReusesProducer(t *testing.T) {
	producer := &fakeProducer{}
	registry := NewKafkaRegistryWithProducer(producer)

	if got := registry.GetKafkaProducer(); got != producer {
		t.Fatalf("expected registry to return the configured producer instance")
	}

	if err := registry.Close(); err != nil {
		t.Fatalf("expected close without error, got %v", err)
	}
	if producer.closeCalls != 1 {
		t.Fatalf("expected producer to be closed once, got %d", producer.closeCalls)
	}
}

func TestRegistryCloseAllowsNilProducer(t *testing.T) {
	registry := &Registry{}

	if err := registry.Close(); err != nil {
		t.Fatalf("expected nil producer close to be a no-op, got %v", err)
	}
}

type closeErrorProducer struct{}

func (closeErrorProducer) ProduceMessage(string, []byte) error {
	return nil
}

func (closeErrorProducer) Close() error {
	return errors.New("close failed")
}

func TestRegistryClosePropagatesProducerError(t *testing.T) {
	registry := NewKafkaRegistryWithProducer(closeErrorProducer{})

	if err := registry.Close(); err == nil {
		t.Fatalf("expected close error")
	}
}
