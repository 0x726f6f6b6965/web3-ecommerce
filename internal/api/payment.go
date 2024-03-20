package api

import (
	"fmt"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api/services"
	"github.com/0x726f6f6b6965/web3-ecommerce/internal/client"
	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/erc20"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

var PaymentApi *paymentApi

type paymentApi struct {
	srv    services.PaymentService
	client *ethclient.Client
}

func NewPaymentApi(serv erc20.ERC20Service, ethClient *ethclient.Client, sqs *client.SQSClient) *paymentApi {
	PaymentApi = &paymentApi{
		srv:    services.NewPaymentService(serv, ethClient, sqs),
		client: ethClient,
	}
	return PaymentApi
}

func (p *paymentApi) Pay(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	pay := new(protos.PayRequest)
	if err := ctx.ShouldBindJSON(pay); err != nil {
		utils.InvalidParamErr.Message = "Please enter correct data."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	if utils.IsEmpty(pay.OrderId) {
		utils.InvalidParamErr.Message = "Please enter correct data."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	nonce, err := p.client.PendingNonceAt(ctx, common.HexToAddress(token.PublicAddress))
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	tx, err := p.srv.PayToken(ctx, token.PublicAddress, pay.OrderId, nonce, pay.Pay)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, tx)
}
