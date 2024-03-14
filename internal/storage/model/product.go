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
	pkProduct = "id"
)

// GetProduct - get product information
// Pk: PRODUCT#<product_id>
// Sk: #PROFILE#<product_id>
func GetProduct(ctx context.Context, client *storage.DaoClient, id string) (*protos.Product, error) {
	info := new(protos.Product)

	data, err := client.DynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(client.Table),
		Key:       storage.GetProductInfoKey(id),
	})

	if err != nil {
		return info, err
	}

	if data.Item == nil {
		return info, storage.ErrNotFound
	}

	if err := attributevalue.UnmarshalMap(data.Item, info); err != nil {
		return info, err
	}

	return info, nil
}

// PutProduct - insert new item.
// Pk: PRODUCT#<product_id>
// Sk: #PROFILE#<product_id>
func PutProduct(ctx context.Context, client *storage.DaoClient, data protos.Product) error {

	item, err := attributevalue.MarshalMap(data)
	if err != nil {
		return err
	}
	item[storage.Sk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(storage.ProfileKey, data.Id),
	}

	_, err = client.DynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(client.Table),
		Item:                item,
		ConditionExpression: aws.String(storage.PkNotExists),
	})
	return err
}

// UpdateProduct - update product information
// Pk: PRODUCT#<product_id>
// Sk: #PROFILE#<product_id>
func UpdateProduct(ctx context.Context, client *storage.DaoClient, id string, info protos.Product, updateMask []string) (*protos.Product, error) {
	newInfo := new(protos.Product)
	expr, err := storage.GetUpdateExpression(info, pkProduct, "", updateMask)
	if err != nil {
		return newInfo, err
	}

	resp, err := client.DynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(client.Table),
		Key:                       storage.GetProductInfoKey(id),
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

// GetAllProducts - get all items
// GSI: soft_deleted_index (soft_deleted)
func GetAllProducts(ctx context.Context, client *storage.DaoClient) ([]*protos.Product, error) {
	var (
		response *dynamodb.ScanOutput
		items    []*protos.Product
	)
	condition := expression.AttributeExists(expression.Name(storage.SoftDeleted))
	expr, err := expression.NewBuilder().WithCondition(condition).Build()
	if err != nil {
		return nil, err
	}

	scanPaginator := dynamodb.NewScanPaginator(client.DynamoClient, &dynamodb.ScanInput{
		TableName:                 aws.String(client.Table),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Condition(),
		IndexName:                 aws.String(storage.SoftDeletedIndex),
	})

	for scanPaginator.HasMorePages() {
		response, err = scanPaginator.NextPage(ctx)
		if err != nil {
			break
		}
		var itemPage []*protos.Product
		err = attributevalue.UnmarshalListOfMaps(response.Items, &itemPage)
		if err != nil {
			break
		}
		items = append(items, itemPage...)
	}
	return items, err
}
