package services

import (
	"context"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage/model"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/google/uuid"
)

type ProductService interface {
	GetProduct(ctx context.Context, id string) (*protos.Product, error)
	CreateProduct(ctx context.Context, product *protos.Product) (*protos.Product, error)
	UpdateProduct(ctx context.Context, id string, product *protos.Product, updateMask []string) (*protos.Product, error)
	GetProducts(ctx context.Context) ([]*protos.Product, error)
}

type productService struct{}

func NewProductService() ProductService {
	return &productService{}
}
func (p *productService) GetProduct(ctx context.Context, id string) (*protos.Product, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	product, err := model.GetProduct(ctx, dynamo, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (p *productService) CreateProduct(ctx context.Context, product *protos.Product) (*protos.Product, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	product.Id = uuid.NewString()
	product.CreatedAt = time.Now().Unix()
	product.UpdatedAt = time.Now().Unix()
	err := model.PutProduct(ctx, dynamo, *product)
	if err != nil {
		return nil, err
	}
	return product, nil
}
func (p *productService) UpdateProduct(ctx context.Context, id string, product *protos.Product, updateMask []string) (*protos.Product, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	newProduct, err := model.UpdateProduct(ctx, dynamo, id, *product, updateMask)
	if err != nil {
		return nil, err
	}
	return newProduct, nil
}

func (p *productService) GetProducts(ctx context.Context) ([]*protos.Product, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	products, err := model.GetAllProducts(ctx, dynamo)
	if err != nil {
		return nil, err
	}
	return products, nil
}
