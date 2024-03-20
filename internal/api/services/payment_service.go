package services

import (
	"context"
	"errors"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/client"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/storage/model"
	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/erc20"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PaymentService interface {
	PayToken(ctx context.Context, publicAddress, orderId string, nonce uint64, in *protos.CommonRequest) (string, error)
}

var (
	rollback uint64 = 5
)

type payment struct {
	token     erc20.ERC20Service
	sqsClient *client.SQSClient
	ether     *ethclient.Client
}

func NewPaymentService(token erc20.ERC20Service, ethClient *ethclient.Client, sqs *client.SQSClient) PaymentService {
	return &payment{
		token:     token,
		ether:     ethClient,
		sqsClient: sqs,
	}
}

func (p *payment) PayToken(ctx context.Context, publicAddress, orderId string, nonce uint64, in *protos.CommonRequest) (string, error) {
	dynamo := storage.GetDynamoClient()
	if dynamo == nil {
		return "", ErrDynamodbClientNotFound
	}

	order, err := model.GetOrder(ctx, dynamo, publicAddress, orderId)

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
		_, err = model.UpdateOrder(ctx, dynamo, publicAddress, orderId,
			*order, []string{"status", "payment_hash", "updated_at"})
		if err != nil {
			// because transaction already done
			// here need to keep the error and update the order
			return "", errors.Join(ErrDynamodb, err)
		}
		// // put tx to sqs
		block, err := p.ether.BlockByNumber(ctx, nil)
		if err != nil {
			// because transaction already done
			return "", errors.Join(ErrEthereum, err)
		}

		sqsData := new(protos.CreateMonitorRequest)
		sqsData.OrderId = orderId
		sqsData.TxHash = tx.Hash().Hex()
		sqsData.Contract = ""
		sqsData.From = publicAddress
		sqsData.FromBlock = block.NumberU64() - rollback
		err = client.Send(ctx, p.sqsClient, sqsData)
		if err != nil {
			// because transaction already done
			// here need to keep the error and send to sqs
			return "", errors.Join(ErrSQS, err)
		}
		return tx.Hash().Hex(), nil
	} else {
		return "", ErrAlreadyPaid
	}
}
