package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"order-service/clients/config"
	"order-service/common/utils"
	configApp "order-service/config"
	"order-service/constants"
	"order-service/domain/dto"
	"time"
)

func NewPaymentClient(client config.IClientConfig) IPaymentClient {
	return &PaymentClient{client: client}
}

type PaymentClient struct {
	client config.IClientConfig
}

func (p *PaymentClient) GetPaymentByUUID(ctx context.Context, uuid uuid.UUID) (*PaymentData, error) {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		configApp.Config.AppName, p.client.SignatureKey(), unixTime)

	apiKey := utils.GenerateSHA256(generateAPIKey)
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	var response PaymentResponse
	request := p.client.Client().Clone().
		Set(constants.Authorization, bearerToken).
		Set(constants.XApiKey, apiKey).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Get(fmt.Sprintf("%s/api/v1/payment/%s", p.client.BaseUrl(), uuid))
	res, _, errs := request.EndStruct(&response)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment response: %", response.Message)
	}
	return &response.Data, nil
}

func (p *PaymentClient) CreatePaymentLink(ctx context.Context, request *dto.PaymentRequest) (*PaymentData, error) {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		configApp.Config.AppName, p.client.SignatureKey(), unixTime)

	apiKey := utils.GenerateSHA256(generateAPIKey)
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	res, bodyRes, errs := p.client.Client().Clone().
		Post(fmt.Sprintf("%s/api/v1/payment", p.client.BaseUrl())).
		Set(constants.Authorization, bearerToken).
		Set(constants.XApiKey, apiKey).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Send(string(body)).
		End()

	if len(errs) > 0 {
		return nil, errs[0]
	}
	var response PaymentResponse
	if res.StatusCode != http.StatusCreated {
		err = json.Unmarshal([]byte(bodyRes), &response)
		if err != nil {
			return nil, err
		}
		paymentErr := fmt.Errorf("payment response: %", response.Message)
		return nil, paymentErr
	}
	err = json.Unmarshal([]byte(bodyRes), &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

type IPaymentClient interface {
	GetPaymentByUUID(context.Context, uuid.UUID) (*PaymentData, error)
	CreatePaymentLink(context.Context, *dto.PaymentRequest) (*PaymentData, error)
}
