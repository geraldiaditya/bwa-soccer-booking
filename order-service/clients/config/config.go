package config

import "github.com/parnurzeal/gorequest"

type ClientConfig struct {
	client       *gorequest.SuperAgent
	baseUrl      string
	signatureKey string
}

type IClientConfig interface {
	Client() *gorequest.SuperAgent
	BaseUrl() string
	SignatureKey() string
}
type Option func(*ClientConfig)

func NewClientConfig(options ...Option) IClientConfig {
	clientConfig := &ClientConfig{
		client: gorequest.New().
			Set("Content-Type", "application/json").
			Set("Accept", "application/json"),
	}

	for _, option := range options {
		option(clientConfig)
	}
	return clientConfig
}

func (cfg *ClientConfig) Client() *gorequest.SuperAgent {
	return cfg.client
}
func (cfg *ClientConfig) BaseUrl() string {
	return cfg.baseUrl
}
func (cfg *ClientConfig) SignatureKey() string {
	return cfg.signatureKey
}

func WithBaseUrl(baseUrl string) Option {
	return func(cfg *ClientConfig) {
		cfg.baseUrl = baseUrl
	}
}
func WithSignatureKey(signatureKey string) Option {
	return func(cfg *ClientConfig) {
		cfg.signatureKey = signatureKey
	}
}
