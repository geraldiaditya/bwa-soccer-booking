package gcs

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"io"
	"time"
)

type ServiceAccountKeyJson struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

type GSClient struct {
	ServiceAccountKeyJson ServiceAccountKeyJson
	BucketName            string
}

type IGSClient interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

func NewGSClient(json ServiceAccountKeyJson, bucketName string) IGSClient {
	return &GSClient{
		ServiceAccountKeyJson: json,
		BucketName:            bucketName,
	}
}

func (g *GSClient) UploadFile(ctx context.Context, filename string, data []byte) (string, error) {
	var (
		contentType      = "application/octet-stream"
		timeoutInSeconds = 60
	)
	client, err := g.createClient(ctx)
	if err != nil {
		logrus.Error("failed to create GS client: %v", err)
		return "", err
	}
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Error("failed to close GS client: %v", err)
			return
		}
	}(client)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()

	bucket := client.Bucket(g.BucketName)
	obj := bucket.Object(filename)
	buffer := bytes.NewBuffer(data)

	writer := obj.NewWriter(ctx)
	writer.ChunkSize = 0
	_, err = io.Copy(writer, buffer)
	if err != nil {
		logrus.Error("failed to copy GS object: %v", err)
		return "", err
	}
	err = writer.Close()
	if err != nil {
		logrus.Error("failed to close Writer: %v", err)
		return "", err
	}
	_, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Error("failed to update : %v", err)
		return "", err
	}
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.BucketName, filename)
	return url, nil
}

func (g *GSClient) createClient(ctx context.Context) (*storage.Client, error) {
	reqBodyBytes := new(bytes.Buffer)
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJson)
	if err != nil {
		logrus.Error("failed to encode service account key :%v", err)
		return nil, err
	}
	jsonByte := reqBodyBytes.Bytes()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonByte))
	if err != nil {
		logrus.Error("failed to create client :%v", err)
		return nil, err
	}
	return client, nil
}
