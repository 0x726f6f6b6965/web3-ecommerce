package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	sqsClient *sqs.Client
	url       string
}

func NewSQSClient(ctx context.Context, region, url string) (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}
	client := new(SQSClient)
	client.sqsClient = sqs.NewFromConfig(cfg)
	client.url = url
	return client, nil
}

func NewDevSQSClient(url, host string, port uint64) *SQSClient {
	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: fmt.Sprintf("http://%s:%d", host, port)}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)
	client := new(SQSClient)
	client.sqsClient = sqs.NewFromConfig(cfg)
	client.url = url
	return client
}

func Send[data *protos.CreateMonitorRequest](ctx context.Context, sqsClient *SQSClient, req data) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = sqsClient.sqsClient.SendMessage(
		ctx,
		&sqs.SendMessageInput{
			QueueUrl:    aws.String(sqsClient.url),
			MessageBody: aws.String(string(b)),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
