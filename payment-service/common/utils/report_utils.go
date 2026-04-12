package utils

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/sirupsen/logrus"
	"html/template"
	"strings"
)

func GeneratePDFFromHTML(htmlTemplate string, data any) ([]byte, error) {
	funcMap := template.FuncMap{
		"add1": add1,
	}

	template, err := template.New("htmlTemplate").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}
	var filledTemplate bytes.Buffer
	if err := template.Execute(&filledTemplate, data); err != nil {
		return nil, err
	}
	htmlContent := filledTemplate.String()
	pdfGenerator, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		logrus.Errorf("failed to create pdf generator: %v", err)
		return nil, err
	}
	pdfGenerator.Dpi.Set(600)
	pdfGenerator.NoCollate.Set(false)
	pdfGenerator.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfGenerator.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfGenerator.Grayscale.Set(false)
	pdfGenerator.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(htmlContent)))

	err = pdfGenerator.Create()
	if err != nil {
		logrus.Errorf("failed to create pdf: %v", err)
		return nil, err
	}
	return pdfGenerator.Bytes(), nil
}

func add1(a int) int {
	return a + 1
}
