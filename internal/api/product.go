package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api/services"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var (
	ProductApi  *productApi
	PRODUCTLIST = "productlist"
)

type productApi struct {
	srv  services.ProductService
	info *cache.Cache
}

func NewProductApi(expire time.Duration) *productApi {
	ProductApi = &productApi{
		srv:  services.NewProductService(),
		info: cache.New(expire, expire*2),
	}
	return ProductApi
}

func (p *productApi) GetProduct(ctx *gin.Context) {
	var productId = ctx.Param("productId")
	if utils.IsEmpty(productId) {
		utils.InvalidParamErr.Message = "Please enter correct productId."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	if PRODUCT, found := p.info.Get(productId); found {
		utils.Response(ctx, utils.SuccessCode, utils.Success, PRODUCT)
		return
	}
	PRODUCT, err := p.srv.GetProduct(ctx, productId)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	p.info.Set(productId, PRODUCT, cache.DefaultExpiration)
	utils.Response(ctx, http.StatusOK, utils.Success, PRODUCT)
}

func (p *productApi) CreateProduct(ctx *gin.Context) {
	var request protos.Product
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.InvalidParamErr.Message = "Please enter correct data."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}

	PRODUCT, err := p.srv.CreateProduct(ctx, &request)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, http.StatusOK, utils.Success, PRODUCT)
}

func (p *productApi) UpdateProduct(ctx *gin.Context) {
	var productId = ctx.Param("productId")
	if utils.IsEmpty(productId) {
		utils.InvalidParamErr.Message = "Please enter correct productId."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}
	var request protos.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.InvalidParamErr.Message = "Please enter correct data."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(request.ProductId) || request.ProductId != productId {
		utils.InvalidParamErr.Message = "Please enter correct product_id."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}

	if request.Product == nil {
		utils.InvalidParamErr.Message = "Please enter correct PRODUCT."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}

	if len(request.UpdateMask) <= 0 {
		utils.InvalidParamErr.Message = "Please enter update mask."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}
	request.Product.Id = productId

	PRODUCT, err := p.srv.UpdateProduct(ctx, productId, request.Product, request.UpdateMask)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, http.StatusOK, utils.Success, PRODUCT)
}

func (p *productApi) GetProductList(ctx *gin.Context) {
	var (
		token = utils.NewPageToken(0, 25)
		err   error
	)
	pageToken := ctx.DefaultQuery("pageToken", "")
	pageSize := ctx.DefaultQuery("pageSize", "0")
	size, err := strconv.ParseUint(pageSize, 10, 64)
	if err != nil {
		utils.InvalidParamErr.Message = "Please enter correct data."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}

	if !utils.IsEmpty(pageToken) {
		token, err = utils.GetPageTokenByString(pageToken)
		if err != nil {
			utils.InvalidParamErr.Message = "Please enter correct page token."
			utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
			return
		}
	}

	if size > 0 {
		token.SetSize(size)
	}

	var (
		resp     protos.GetProductListResponse
		start    = token.GetID()
		end      = start + int(token.GetSize())
		PRODUCTs = make([]*protos.Product, 0)
	)

	if products, found := p.info.Get(PRODUCTLIST); found {
		PRODUCTs = products.([]*protos.Product)
	} else {
		PRODUCTs, err = p.srv.GetProducts(ctx)
		if err != nil {
			utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
			utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
			return
		}
		p.info.Set(PRODUCTLIST, PRODUCTs, cache.DefaultExpiration)
	}

	if end > len(PRODUCTs) {
		resp.NextPageToken = ""
		resp.Products = PRODUCTs[start:]
	} else {
		resp.Products = PRODUCTs[start:end]
		token.SetID(end)
		resp.NextPageToken = token.GetToken()
	}
	utils.Response(ctx, http.StatusOK, utils.Success, resp)
}
