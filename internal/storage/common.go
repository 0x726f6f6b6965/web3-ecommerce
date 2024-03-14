package storage

import (
	"fmt"
	"reflect"

	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetUserInfoKey(address string) map[string]types.AttributeValue {
	result := make(map[string]types.AttributeValue)
	result[Pk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(UserKey, address),
	}
	result[Sk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(ProfileKey, address),
	}
	return result
}

func GetUserOrderKey(address, orderId string) map[string]types.AttributeValue {
	result := make(map[string]types.AttributeValue)
	result[Pk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(UserKey, address),
	}
	result[Sk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(OrderKey, orderId),
	}
	return result
}

func GetProductInfoKey(id string) map[string]types.AttributeValue {
	result := make(map[string]types.AttributeValue)
	result[Pk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(ProductKey, id),
	}
	result[Sk] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf(ProfileKey, id),
	}
	return result
}

func GetUpdateExpression(in interface{}, pk, sk string, updateMask []string) (expression.Expression, error) {
	var (
		vals   = reflect.ValueOf(in)
		start  = true
		update expression.UpdateBuilder
	)
	for _, key := range updateMask {
		sKey := utils.ToCamelCase(key)
		if sKey == pk || sKey == sk {
			continue
		}

		if vals.FieldByName(sKey).IsValid() {
			if start {
				update = expression.Set(expression.Name(key), expression.Value(vals.FieldByName(sKey).Interface()))
				start = false
			} else {
				update.Set(expression.Name(key), expression.Value(vals.FieldByName(sKey).Interface()))
			}
		}
	}
	return expression.NewBuilder().WithUpdate(update).Build()
}
