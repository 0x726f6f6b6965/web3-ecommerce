package services

import (
	"context"
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage/model"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *protos.Order) (*protos.Order, error)
	GetOrder(ctx context.Context, publicAddress, id string) (*protos.Order, error)
	GetUserOrder(ctx context.Context, publicAddress string) ([]protos.Order, error)
	UpdateOrder(ctx context.Context, publicAddress, id string, order *protos.Order, updateMask []string) error
}

type orderService struct {
}

func NewOrderService() OrderService {
	return &orderService{}
}

func (s *orderService) CreateOrder(ctx context.Context, order *protos.Order) (*protos.Order, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	_, err := model.GetUserInfo(ctx, dynamo, order.From)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	order.Id = id
	order.Status = protos.StatusCreated
	order.CreatedAt = time.Now().Unix()
	order.UpdatedAt = time.Now().Unix()
	order.StatusCreatedAt = fmt.Sprintf("%s#%d", order.Status, order.CreatedAt)
	err = model.PutOrder(ctx, dynamo, *order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, publicAddress, id string) (*protos.Order, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	order, err := model.GetOrder(ctx, dynamo, publicAddress, id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetUserOrder(ctx context.Context, publicAddress string) ([]protos.Order, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	orders, err := model.GetUserOrders(ctx, dynamo, publicAddress)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *orderService) UpdateOrder(ctx context.Context, publicAddress, id string, order *protos.Order, updateMask []string) error {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return ErrDynamodbClientNotFound
	}
	_, err := model.UpdateOrder(ctx, dynamo, publicAddress, id, *order, updateMask)
	if err != nil {
		return err
	}
	return nil
}
