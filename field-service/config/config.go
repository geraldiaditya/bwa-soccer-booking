package config

import (
	"errors"
	"field-service/common/utils"
	"os"

	"github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
)

var Config AppConfig

type AppConfig struct {
	Port                       int             `json:"port"`
	AppName                    string          `json:"appName"`
	AppEnv                     string          `json:"appEnv"`
	SignatureKey               string          `json:"signatureKey"`
	Database                   Database        `json:"database"`
	RateLimiterMaxRequest      float64         `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond      float64         `json:"rateLimiterTimeSecond"`
	InternalService            InternalService `json:"internalService"`
	GCSType                    string          `json:"gcsType"`
	GCSProjectID               string          `json:"gcsProjectID"`
	GCSPrivateKeyID            string          `json:"gcsPrivateKeyID"`
	GCSPrivateKey              string          `json:"gcsPrivateKey"`
	GCSClientEmail             string          `json:"gcsClientEmail"`
	GCSClientID                string          `json:"gcsClientID"`
	GCSAuthURI                 string          `json:"gcsAuthURI"`
	GCSTokenURI                string          `json:"gcsTokenURI"`
	GCSAuthProviderX509CertURL string          `json:"gcsAuthProviderX509CertURL"`
	GCSClientX509CertURL       string          `json:"gcsClientX509CertURL"`
	GCSUniverseDomain          string          `json:"gcsUniverseDomain"`
	GCSBucketName              string          `json:"gcsBucketName"`
	AllowedOrigins             []string        `json:"allowedOrigins"`
}

type Database struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	Name                  string `json:"name"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	MaxOpenConnection     int    `json:"maxOpenConnection"`
	MaxLifetimeConnection int    `json:"maxLifetimeConnection"`
	MaxIdleConnection     int    `json:"maxIdleConnection"`
	MaxIdleTime           int    `json:"maxIdleTime"`
}

type InternalService struct {
	User User `json:"user"`
}

type User struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

func (c *AppConfig) Validate() error {
	if c.Port <= 0 {
		return errors.New("config: port is required and must be > 0")
	}
	if c.Database.Host == "" {
		return errors.New("config: database host is required")
	}
	if c.Database.Name == "" {
		return errors.New("config: database name is required")
	}
	if c.Database.Username == "" {
		return errors.New("config: database username is required")
	}
	if c.SignatureKey == "" {
		return errors.New("config: signatureKey is required")
	}
	return nil
}

func Init() {
	err := utils.BindFromJSON(&Config, "config.json", ".")
	if err != nil {
		logrus.Infof("failed to bind config json:%v", err)
		consulURL := os.Getenv("CONSUL_HTTP_URL")
		consulKey := os.Getenv("CONSUL_HTTP_KEY")
		if consulURL != "" && consulKey != "" {
			err = utils.BindFromConsul(&Config, consulURL, consulKey)
			if err != nil {
				logrus.Errorf("failed to bind config from consul: %v", err)
			}
		}
	}
}
