package config

import (
	"errors"
	"order-service/common/utils"
	"os"

	"github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
)

var Config AppConfig

type AppConfig struct {
	Port                  int             `json:"port"`
	AppName               string          `json:"appName"`
	AppEnv                string          `json:"appEnv"`
	SignatureKey          string          `json:"signatureKey"`
	Database              Database        `json:"database"`
	RateLimiterMaxRequest float64         `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond float64         `json:"rateLimiterTimeSecond"`
	AllowedOrigins        []string        `json:"allowedOrigins"`
	InternalService       InternalService `json:"internalService"`
	//GCSType                    string          `json:"gcsType"`
	//GCSProjectID               string          `json:"gcsProjectID"`
	//GCSPrivateKeyID            string          `json:"gcsPrivateKeyID"`
	//GCSPrivateKey              string          `json:"gcsPrivateKey"`
	//GCSClientEmail             string          `json:"gcsClientEmail"`
	//GCSClientID                string          `json:"gcsClientID"`
	//GCSAuthURI                 string          `json:"gcsAuthURI"`
	//GCSTokenURI                string          `json:"gcsTokenURI"`
	//GCSAuthProviderX509CertURL string          `json:"gcsAuthProviderX509CertURL"`
	//GCSClientX509CertURL       string          `json:"gcsClientX509CertURL"`
	//GCSUniverseDomain          string          `json:"gcsUniverseDomain"`
	//GCSBucketName              string          `json:"gcsBucketName"`
	Kafka Kafka `json:"kafka"`
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
	User    User    `json:"user"`
	Field   Field   `json:"field"`
	Payment Payment `json:"payment"`
}

type User struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Field struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Payment struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Kafka struct {
	Brokers               []string `json:"brokers"`
	TimeoutInMS           int      `json:"timeoutInMS"`
	MaxRetry              int      `json:"maxRetry"`
	Topics                []string `json:"topics"`
	GroupID               string   `json:"groupId"`
	MaxWaitTimeInMs       int      `json:"maxWaitTimeInMs"`
	MaxProcessingTimeInMs int      `json:"maxProcessingTimeInMs"`
	BackoffTimeInMS       int      `json:"backoffTimeInMS"`
}

type Midtrans struct {
	ServerKey    string `json:"serverKey"`
	ClientKey    string `json:"clientKey"`
	IsProduction bool   `json:"isProduction"`
}

func Init() {
	err := utils.BindFromJSON(&Config, "config.json", ".")
	if err != nil {
		logrus.Infof("failed to bind config json:%v", err)
		err = utils.BindFromConsul(&Config, os.Getenv("CONSUL_HTTP_URL"), os.Getenv("CONSUL_HTTP_key"))
		if err != nil {
			panic(err)
		}
	}
}

func (c *AppConfig) Validate() error {
	if c.Port == 0 {
		return errors.New("config: port is required")
	}
	if c.AppName == "" {
		return errors.New("config: appName is required")
	}
	if c.SignatureKey == "" {
		return errors.New("config: signatureKey is required")
	}
	if c.InternalService.User.Host == "" {
		return errors.New("config: internalService.user.host is required")
	}
	if c.InternalService.Field.Host == "" {
		return errors.New("config: internalService.field.host is required")
	}
	if c.InternalService.Payment.Host == "" {
		return errors.New("config: internalService.payment.host is required")
	}
	if len(c.Kafka.Brokers) == 0 {
		return errors.New("config: kafka.brokers is required")
	}
	return nil
}
