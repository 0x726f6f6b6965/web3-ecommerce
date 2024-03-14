package model

import (
	"context"
	"fmt"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	pkUser = "public_address"
)

// GetUserInfo - get user information
// Pk: USER#<public address>
// Sk: #PROFILE#<public address>
func GetUserInfo(ctx context.Context, client *storage.DaoClient, publicAddress string) (*protos.User, error) {
	info := new(protos.User)

	data, err := client.DynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(client.Table),
		Key:       storage.GetUserInfoKey(publicAddress),
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

// PutUserInfo - put user information
// PK: USER#<public address>
// SK: #PROFILE#<public address>
func PutUserInfo(ctx context.Context, client *storage.DaoClient, info protos.User) error {
	data, err := attributevalue.MarshalMap(info)
	if err != nil {
		return err
	}
	data[storage.Pk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(storage.UserKey, info.PublicAddress),
	}

	data[storage.Sk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(storage.ProfileKey, info.PublicAddress),
	}

	_, err = client.DynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(client.Table),
		Item:                data,
		ConditionExpression: aws.String(storage.PkNotExists),
	})

	return err
}

// DeleteUserInfo - delete user information
// PK: USER#<public address>
// SK: #PROFILE#<public address>
func DeleteUserInfo(ctx context.Context, client *storage.DaoClient, publicAddress string) error {
	_, err := client.DynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           aws.String(client.Table),
		Key:                 storage.GetUserInfoKey(publicAddress),
		ConditionExpression: aws.String(storage.PkExists),
	})

	return err
}

// UpdateUserInfo - update user information
// PK: USER#<public address>
// SK: #PROFILE#<public address>
func UpdateUserInfo(ctx context.Context, client *storage.DaoClient, publicAddress string, info protos.User, updateMask []string) (*protos.User, error) {
	newInfo := new(protos.User)
	expr, err := storage.GetUpdateExpression(info, pkUser, "", updateMask)
	if err != nil {
		return newInfo, err
	}

	resp, err := client.DynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(client.Table),
		Key:                       storage.GetUserInfoKey(publicAddress),
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
