package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/helper"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage/model"
	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/erc20"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type UserService interface {
	GetToken(ctx context.Context, publicAddress string, signature string) (string, error)
	GetUserInfo(ctx context.Context, publicAddress string) (*protos.User, error)
	UpdateUserInfo(ctx context.Context, publicAddress string, user *protos.User, updateMask []string) (*protos.User, error)
	CreateUser(ctx context.Context, user *protos.User) error
}

var (
	LoginMsg   = "%s Login! Nonce: %d"
	ExpireTime = 5 * time.Minute
)

type userService struct {
	client *ethclient.Client
}

func NewUserService(client *ethclient.Client) UserService {
	return &userService{
		client: client,
	}
}

func (s *userService) GetToken(ctx context.Context, publicAddress string, signature string) (string, error) {
	nonce, err := s.client.PendingNonceAt(ctx, common.HexToAddress(publicAddress))
	if err != nil {
		return "", errors.Join(erc20.ErrEthClient, err)
	}
	msg := fmt.Sprintf(LoginMsg, publicAddress, nonce)
	err = utils.VerifySignature(publicAddress, signature, msg)
	if err != nil {
		return "", errors.Join(ErrInvalidSignature, err)
	}
	token, err := helper.GenerateNewAccessToken(publicAddress, nonce, ExpireTime)
	if err != nil {
		return "", errors.Join(ErrGenerateToken, err)
	}
	return token, nil
}

func (s *userService) GetUserInfo(ctx context.Context, publicAddress string) (*protos.User, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	user, err := model.GetUserInfo(ctx, dynamo, publicAddress)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUserInfo(ctx context.Context, publicAddress string, user *protos.User, updateMask []string) (*protos.User, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return nil, ErrDynamodbClientNotFound
	}
	user, err := model.UpdateUserInfo(ctx, dynamo, publicAddress, *user, updateMask)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) CreateUser(ctx context.Context, user *protos.User) error {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return ErrDynamodbClientNotFound
	}
	user.CreatedAt = time.Now().Unix()
	user.UpdatedAt = time.Now().Unix()
	err := model.PutUserInfo(ctx, dynamo, *user)
	if err != nil {
		return err
	}
	return nil
}
