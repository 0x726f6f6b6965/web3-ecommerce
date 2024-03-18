package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage/model"
	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/erc20"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PaymentService interface {
	PayToken(ctx context.Context, publicAddress, orderId string, nonce uint64, in *protos.CommonRequest) (string, error)
}

type payment struct {
	token erc20.ERC20Service
}

func NewPaymentService(token erc20.ERC20Service, client *ethclient.Client) PaymentService {
	return &payment{
		token: token,
	}
}

func (p *payment) PayToken(ctx context.Context, publicAddress, orderId string, nonce uint64, in *protos.CommonRequest) (string, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return "", ErrDynamodbClientNotFound
	}

	order, err := model.GetOrder(ctx, dynamo,
		fmt.Sprintf(storage.UserKey, publicAddress),
		fmt.Sprintf(storage.OrderKey, orderId))

	if err != nil {
		return "", err
	}

	if order.Amount != in.Amount {
		return "", ErrInvalidAmount
	}

	if nonce != in.Nonce {
		return "", erc20.ErrInvalidNonce
	}

	if order.Status == protos.StatusCreated || order.Status == protos.StatusPaidFailed {
		tx, err := p.token.TransferWithSign(ctx, *in)
		if err != nil {
			return "", errors.Join(ErrTransactionFailed, err)
		}
		order.Status = protos.StatusPending
		order.PaymentHash = tx.Hash().Hex()
		order.UpdatedAt = time.Now().Unix()
		_, err = model.UpdateOrder(ctx, dynamo,
			fmt.Sprintf(storage.UserKey, publicAddress),
			fmt.Sprintf(storage.OrderKey, orderId),
			*order, []string{"status", "payment_hash", "updated_at"})
		if err != nil {
			return "", errors.Join(ErrDynamodb, err)
		}
		return tx.Hash().Hex(), nil
	} else {
		return "", ErrAlreadyPaid
	}
}
