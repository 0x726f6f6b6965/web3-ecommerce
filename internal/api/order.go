package api

import (
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api/services"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/gin-gonic/gin"
)

var OrderApi *orderApi

type orderApi struct {
	srv     services.OrderService
	product services.ProductService
}

func NewOrderApi() *orderApi {
	OrderApi = &orderApi{
		srv:     services.NewOrderService(),
		product: services.NewProductService(),
	}
	return OrderApi
}

func (o *orderApi) GetOrder(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	var orderId = ctx.Param("orderId")
	if utils.IsEmpty(orderId) {
		utils.InvalidParamErr.Message = "Please enter correct orderId."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	data, err := o.srv.GetOrder(ctx, token.PublicAddress, orderId)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, data)
}

func (o *orderApi) GetOrders(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	data, err := o.srv.GetUserOrder(ctx, token.PublicAddress)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, data)
}

func (o *orderApi) CreateOrder(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	order := new(protos.Order)
	if err := ctx.ShouldBindJSON(order); err != nil {
		utils.InvalidParamErr.Message = "Please enter correct data."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(order.From) ||
		!utils.IsValidAddress(order.From) ||
		token.PublicAddress != order.From {
		utils.InvalidParamErr.Message = "Please enter correct from."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if len(order.ProductIds) == 0 {
		utils.InvalidParamErr.Message = "Please enter products."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if order.Address == "" {
		utils.InvalidParamErr.Message = "Please enter address."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	var total float64
	for _, product := range order.ProductIds {
		info, err := o.product.GetProduct(ctx, product.Id)
		if err != nil {
			utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
			utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
			return
		}
		total += (info.Price * float64(product.Quantity))
	}
	if total != order.Amount {
		utils.InvalidParamErr.Message = "Please enter correct total."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	order, err = o.srv.CreateOrder(ctx, order)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, order.Id)
}

func (o *orderApi) CancelOrder(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	var orderId = ctx.Param("orderId")
	if utils.IsEmpty(orderId) {
		utils.InvalidParamErr.Message = "Please enter correct orderId."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	order, err := o.srv.GetOrder(ctx, token.PublicAddress, orderId)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}

	order.Id = orderId
	order.Status = protos.StatusCancelled
	order.From = token.PublicAddress
	order.UpdatedAt = time.Now().Unix()
	order.StatusCreatedAt = fmt.Sprintf("%s#%d", protos.StatusCancelled.String(), order.CreatedAt)
	mask := []string{"updated_at", "status", "status_created_at"}
	if err := o.srv.UpdateOrder(ctx, token.PublicAddress, orderId, order, mask); err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, nil)
}

func (o *orderApi) UpdateOrderStatus(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	var param protos.UpdateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.InvalidParamErr.Message = err.Error()
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	if utils.IsEmpty(param.OrderId) {
		utils.InvalidParamErr.Message = "Please enter order id."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	order, err := o.srv.GetOrder(ctx, token.PublicAddress, param.OrderId)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	order.Id = param.OrderId
	order.Status = param.Status
	order.UpdatedAt = time.Now().Unix()
	order.StatusCreatedAt = fmt.Sprintf("%s#%d", param.Status.String(), order.CreatedAt)
	mask := []string{"updated_at", "status", "status_created_at"}
	if err := o.srv.UpdateOrder(ctx, token.PublicAddress, param.OrderId, order, mask); err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, nil)
}
