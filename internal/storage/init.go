package storage

import (
	"context"
	"fmt"

	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/once"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	onceInit  once.Once
	daoClient *DaoClient
)

type DaoClient struct {
	DynamoClient *dynamodb.Client
	Table        string
}

func NewDynamoClient(ctx context.Context, region, table string) error {
	var (
		err    error
		cfg    aws.Config
		client *DaoClient
	)
	// singleton
	onceInit.Do(func() error {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err == nil {
			client.Table = table
			client.DynamoClient = dynamodb.NewFromConfig(cfg)
			daoClient = client
		}
		return err
	})
	return err
}

func NewDevLocalClient(table string, host string, port uint64) error {
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
	dynamo := dynamodb.NewFromConfig(cfg)
	client := &DaoClient{
		Table:        table,
		DynamoClient: dynamo,
	}
	daoClient = client
	return nil
}

func GetDynamoClient() *DaoClient {
	return daoClient
}
