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

func NewFieldClient(client config.IClientConfig) IFieldClient {
	return &FieldClient{client: client}
}

type FieldClient struct {
	client config.IClientConfig
}

func (f *FieldClient) GetFieldByUUID(ctx context.Context, uuid uuid.UUID) (*FieldData, error) {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		configApp.Config.AppName, f.client.SignatureKey(), unixTime)

	apiKey := utils.GenerateSHA256(generateAPIKey)
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	var response FieldResponse
	request := f.client.Client().Clone().
		Set(constants.Authorization, bearerToken).
		Set(constants.XApiKey, apiKey).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Get(fmt.Sprintf("%s/api/v1/field/schedule/%s", f.client.BaseUrl(), uuid))
	res, _, errs := request.EndStruct(&response)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("field response: %", response.Message)
	}
	return &response.Data, nil
}

func (f *FieldClient) UpdateStatus(request *dto.UpdateStatusFieldScheduleRequest) error {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		configApp.Config.AppName, f.client.SignatureKey(), unixTime)

	apiKey := utils.GenerateSHA256(generateAPIKey)

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	res, bodyRes, errs := f.client.Client().Clone().
		Post(fmt.Sprintf("%s/api/v1/field/schedule/status", f.client.BaseUrl())).
		Set(constants.XApiKey, apiKey).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Send(string(body)).
		End()

	if len(errs) > 0 {
		return errs[0]
	}
	var response FieldResponse
	if res.StatusCode != http.StatusCreated {
		err = json.Unmarshal([]byte(bodyRes), &response)
		if err != nil {
			return err
		}
		fieldErr := fmt.Errorf("field response: %", response.Message)
		return fieldErr
	}
	err = json.Unmarshal([]byte(bodyRes), &response)
	if err != nil {
		return err
	}
	return nil
}

type IFieldClient interface {
	GetFieldByUUID(context.Context, uuid.UUID) (*FieldData, error)
	UpdateStatus(request *dto.UpdateStatusFieldScheduleRequest) error
}
