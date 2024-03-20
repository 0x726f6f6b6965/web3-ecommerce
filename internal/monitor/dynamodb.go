package monitor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

const (
	PkExists string = "attribute_exists(pk)"
)

var (
	ErrExpression error = errors.New("expression error")
	ErrUpdate     error = errors.New("update error")
)

// UpdateOrder - update order.
// Pk: USER#<public address>
// Sk: ORDER#<order_id>
func UpdateTransStatus(ctx context.Context, client *dynamodb.Client, data *protos.UpdateTrans) error {
	keys := make(map[string]types.AttributeValue)
	keys["pk"] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf("USER#%s", data.From),
	}
	keys["sk"] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf("ORDER#%s", data.OrderId),
	}

	// expression
	now := time.Now().Unix()
	update := expression.Set(expression.Name("payment_hash"), expression.Value(data.TxHash))
	update.Set(expression.Name("status"), expression.Value(data.Status))
	update.Set(expression.Name("updated_at"), expression.Value(now))
	update.Set(expression.Name("status_created_at"),
		expression.Value(fmt.Sprintf("%s#%d", data.Status.String(), now)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return errors.Join(ErrExpression, err)
	}

	_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(data.Table),
		Key:                       keys,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueNone,
		ConditionExpression:       aws.String(PkExists),
	})
	if err != nil {
		return errors.Join(ErrUpdate, err)
	}
	return nil
}
