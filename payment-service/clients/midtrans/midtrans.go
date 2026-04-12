package clients

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sirupsen/logrus"
	errPayment "payment-service/constants/error/payment"
	"payment-service/domain/dto"
	"time"
)

type IMidTransClient interface {
	CreatePaymentLink(response *dto.PaymentRequest) (*MidTransData, error)
}

func NewMidTransClient(serverKey string, isProduction bool) IMidTransClient {
	return &MidTransClient{
		ServerKey:    serverKey,
		IsProduction: isProduction,
	}
}

type MidTransClient struct {
	ServerKey    string
	IsProduction bool
}

func (m *MidTransClient) CreatePaymentLink(req *dto.PaymentRequest) (*MidTransData, error) {
	var (
		snapClient   snap.Client
		isProduction = midtrans.Sandbox
	)

	expiryDateTime := req.ExpiredAt
	currentTime := time.Now()

	duration := expiryDateTime.Sub(currentTime)
	if duration <= 0 {
		logrus.Errorf("Expired at is invalid")
		return nil, errPayment.ErrExpireAtInvalid
	}
	expiryUnit := "minute"
	expiryDuration := int64(duration.Minutes())
	if duration.Hours() >= 1 {
		expiryUnit = "hour"
		expiryDuration = int64(duration.Hours())
	} else if duration.Hours() >= 24 {
		expiryUnit = "day"
		expiryDuration = int64(duration.Hours() / 24)
	}

	if m.IsProduction {
		isProduction = midtrans.Production
	}
	snapClient.New(m.ServerKey, isProduction)
	midRequest := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: int64(req.Amount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.CustomerDetail.Name,
			Email: req.CustomerDetail.Email,
			Phone: req.CustomerDetail.Phone,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    req.ItemDetails[0].ID,
				Name:  req.ItemDetails[0].Name,
				Price: int64(req.ItemDetails[0].Amount),
				Qty:   int32(req.ItemDetails[0].Quantity),
			},
		},
		Expiry: &snap.ExpiryDetails{
			Unit:     expiryUnit,
			Duration: expiryDuration,
		},
	}
	res, err := snapClient.CreateTransaction(midRequest)
	if err != nil {
		logrus.Errorf("Error Create Transaction: %v", err)
		return nil, err
	}
	return &MidTransData{
		RedirectURL: res.RedirectURL,
		Token:       res.Token,
	}, nil
}
