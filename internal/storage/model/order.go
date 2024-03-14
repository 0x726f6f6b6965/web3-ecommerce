package model

import (
	"context"
	"fmt"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	pkOrder = "id"
	skOrder = "from"
)

// GetUserOrders - get all orders of user
// PK: USER#<public address>
// SK: BeginWith ORDER#
func GetUserOrders(ctx context.Context, client *storage.DaoClient, publicAddress string) ([]protos.Order, error) {
	var (
		response *dynamodb.QueryOutput
		orders   []protos.Order
	)

	keyEx := expression.KeyAnd(
		expression.Key(storage.Pk).Equal(expression.Value(fmt.Sprintf(storage.UserKey, publicAddress))),
		expression.KeyBeginsWith(expression.Key(storage.Sk), fmt.Sprintf(storage.OrderKey, "")))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return orders, err
	}
	queryPaginator := dynamodb.NewQueryPaginator(client.DynamoClient, &dynamodb.QueryInput{
		TableName:                 aws.String(client.Table),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		AttributesToGet:           []string{storage.Pk, storage.Sk, "status", "created_at"},
	})

	for queryPaginator.HasMorePages() {
		response, err = queryPaginator.NextPage(ctx)
		if err != nil {
			break
		}
		var orderPage []protos.Order
		err = attributevalue.UnmarshalListOfMaps(response.Items, &orderPage)
		if err != nil {
			break
		}
		orders = append(orders, orderPage...)
	}
	return orders, err
}

// PutOrder - insert new order.
// Pk: USER#<public address>
// Sk: ORDER#<order_id>
func PutOrder(ctx context.Context, client *storage.DaoClient, order protos.Order) error {
	item, err := attributevalue.MarshalMap(order)
	if err != nil {
		return err
	}

	_, err = client.DynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(client.Table),
		Item:                item,
		ConditionExpression: aws.String(storage.PkNotExists),
	})
	return err
}

// GetOrder - get order by order id.
// Pk: USER#<public address>
// Sk: ORDER#<order_id>
func GetOrder(ctx context.Context, client *storage.DaoClient, publicAddress string, orderId string) (*protos.Order, error) {
	order := new(protos.Order)

	response, err := client.DynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(client.Table),
		Key:       storage.GetUserOrderKey(publicAddress, orderId),
	})
	if err != nil {
		return order, err
	}
	if response.Item == nil {
		return order, nil
	}
	err = attributevalue.UnmarshalMap(response.Item, &order)
	return order, err
}

// UpdateOrder - update order.
// Pk: USER#<public address>
// Sk: ORDER#<order_id>
func UpdateOrder(ctx context.Context, client *storage.DaoClient, publicAddress, orderId string, order protos.Order, updateMask []string) (*protos.Order, error) {
	newInfo := new(protos.Order)
	expr, err := storage.GetUpdateExpression(order, pkOrder, skOrder, updateMask)
	if err != nil {
		return newInfo, err
	}

	resp, err := client.DynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(client.Table),
		Key:                       storage.GetUserOrderKey(publicAddress, orderId),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueAllNew,
		ConditionExpression:       aws.String(storage.PkExists),
	})

	if err != nil {
		return newInfo, err
	}

	err = attributevalue.UnmarshalMap(resp.Attributes, newInfo)
	if err != nil {
		return newInfo, err
	}

	return newInfo, nil
}
