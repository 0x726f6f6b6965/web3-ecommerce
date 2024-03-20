package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/monitor"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ether *ethclient.Client
	db    *dynamodb.Client
)
var TimeOut = time.Minute * 3

func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	if len(sqsEvent.Records) < 1 {
		return errors.Join(ErrInvalidEvent, errors.New("the event is empty"))
	}
	request := new(protos.CreateMonitorRequest)
	err := json.Unmarshal([]byte(sqsEvent.Records[0].Body), request)
	if err != nil {
		return errors.Join(ErrUnmarshal, err)
	}

	data, stop, errChan := monitor.Monitor(ether, request)
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
	lambda.Start(Handler)
}
