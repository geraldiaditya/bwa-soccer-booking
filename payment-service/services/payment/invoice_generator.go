package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"payment-service/common/gcs"
	"payment-service/common/utils"
	"payment-service/domain/dto"
	"payment-service/domain/models"
)

const invoiceTemplatePath = "templates/invoice.html"

type pdfRenderer func(*dto.InvoiceRequest) ([]byte, error)

type InvoiceGenerator struct {
	gcs              gcs.IGSClient
	htmlTemplatePath string
	renderPDF        pdfRenderer
}

func NewInvoiceGenerator(gcs gcs.IGSClient) *InvoiceGenerator {
	generator := &InvoiceGenerator{
		gcs:              gcs,
		htmlTemplatePath: invoiceTemplatePath,
	}
	generator.renderPDF = generator.generatePDF
	return generator
}

func BuildInvoiceRequest(payment *models.Payment, webhook *dto.Webhook, paidAt time.Time, invoiceNumber string) *dto.InvoiceRequest {
	total := utils.RupiahFormat(&payment.Amount)
	return &dto.InvoiceRequest{
		InvoiceNumber: invoiceNumber,
		Data: dto.InvoiceData{
			PaymentDetail: dto.InvoicePaymentDetail{
				BankName:      strings.ToUpper(valueOrEmpty(payment.Bank)),
				PaymentMethod: webhook.PaymentType,
				VANumber:      valueOrEmpty(payment.VANumber),
				Date:          utils.FormatIndonesianDate(paidAt),
				IsPaid:        true,
			},
			Items: []dto.InvoiceItem{
				{
					Description: valueOrEmpty(payment.Description),
					Price:       total,
				},
			},
			Total: total,
		},
	}
}

func BuildInvoiceNumber(now time.Time, orderID uuid.UUID, paymentID uuid.UUID) string {
	return fmt.Sprintf("INV/%s/ORD/%s-%s", now.Format(time.DateOnly), uuidSegment(orderID), uuidSegment(paymentID))
}

func InvoiceFilename(invoiceNumber string) string {
	invoiceNumberReplace := strings.ToLower(strings.ReplaceAll(invoiceNumber, "/", "-"))
	return fmt.Sprintf("%s.pdf", invoiceNumberReplace)
}

func (g *InvoiceGenerator) GenerateAndUpload(ctx context.Context, req *dto.InvoiceRequest) (string, error) {
	pdf, err := g.renderPDF(req)
	if err != nil {
		return "", err
	}
	return g.uploadToGCS(ctx, req.InvoiceNumber, pdf)
}

func (g *InvoiceGenerator) generatePDF(req *dto.InvoiceRequest) ([]byte, error) {
	htmlTemplate, err := os.ReadFile(g.htmlTemplatePath)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	jsonData, _ := json.Marshal(req)
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	return utils.GeneratePDFFromHTML(string(htmlTemplate), data)
}

func (g *InvoiceGenerator) uploadToGCS(ctx context.Context, invoiceNumber string, pdf []byte) (string, error) {
	return g.gcs.UploadFile(ctx, InvoiceFilename(invoiceNumber), pdf)
}

func uuidSegment(id uuid.UUID) string {
	return id.String()[:8]
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
