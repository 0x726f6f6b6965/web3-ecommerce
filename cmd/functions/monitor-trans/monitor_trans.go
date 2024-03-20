package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/monitor"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var (
	Ether *ethclient.Client
	db    *dynamodb.Client
)
var (
	TimeOut               = time.Minute * 3
	ErrInvalidEvent error = errors.New("invalid event")
	ErrUnmarshal    error = errors.New("unmarshal error")
	ErrTimeout      error = errors.New("timeout")
	ErrMonitor      error = errors.New("monitor error")
	ErrUpdateTrans  error = errors.New("update transaction error")
)

func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	if len(sqsEvent.Records) < 1 {
		return errors.Join(ErrInvalidEvent, errors.New("the event is empty"))
	}
	request := new(protos.CreateMonitorRequest)
	err := json.Unmarshal([]byte(sqsEvent.Records[0].Body), request)
	if err != nil {
		return errors.Join(ErrUnmarshal, err)
	}

	data, stop, errChan := monitor.Monitor(Ether, request)
	ctx, cancel := context.WithTimeout(ctx, TimeOut)
	defer cancel()

	trans := new(protos.UpdateTrans)
	trans.From = request.From
	trans.OrderId = request.OrderId
	trans.TxHash = request.TxHash
	trans.Table = request.Table

	select {
	case <-ctx.Done():
		stop()
		trans.Status = protos.StatusMonitorFailed
		if err := monitor.UpdateTransStatus(ctx, db, trans); err != nil {
			return errors.Join(ErrUpdateTrans, err)
		}
		return errors.Join(ErrTimeout, ctx.Err())
	case err := <-errChan:
		stop()
		trans.Status = protos.StatusMonitorFailed
		if dberr := monitor.UpdateTransStatus(ctx, db, trans); dberr != nil {
			return errors.Join(ErrUpdateTrans, dberr)
		}
		return errors.Join(ErrMonitor, err)
	case vLog := <-data:
		if len(vLog.Topics) != 3 {
			return errors.Join(ErrInvalidEvent, errors.New("the event topics error"))
		}
		trans.Status = protos.StatusPaid
		if err := monitor.UpdateTransStatus(ctx, db, trans); err != nil {
			return errors.Join(ErrUpdateTrans, err)
		}
		return nil
	}
}
func main() {
	godotenv.Load()
	client, err := ethclient.Dial(os.Getenv("RPC"))
	if err != nil {
		panic(err)
	}
	Ether = client
	var cfg aws.Config
	if os.Getenv("ENV") == "dev" {
		cfg, _ = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion("us-east-1"),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: fmt.Sprintf("http://%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))}, nil
				})))
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(os.Getenv("REGION")),
		)
		if err != nil {
			panic(err)
		}
	}
	db = dynamodb.NewFromConfig(cfg)

	lambda.Start(Handler)
}
