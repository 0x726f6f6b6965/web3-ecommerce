package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/api/services"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

var UserApi *userApi

type userApi struct {
	srv services.UserService
}

func NewUserApi(client *ethclient.Client) *userApi {
	UserApi = &userApi{srv: services.NewUserService(client)}
	return UserApi
}

func (u *userApi) GetUser(ctx *gin.Context) {
	userToken, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	user, err := u.srv.GetUserInfo(ctx, userToken.PublicAddress)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, user)
}

func (u *userApi) GetToken(ctx *gin.Context) {
	var param protos.GetTokenRequest
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.InvalidParamErr.Message = err.Error()
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(param.PublicAddress) || !utils.IsValidAddress(param.PublicAddress) {
		utils.InvalidParamErr.Message = "Please enter correct address."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(param.Signature) {
		utils.InvalidParamErr.Message = "Please enter signature."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if token, err := u.srv.GetToken(ctx, param.PublicAddress, param.Signature); err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
	} else {
		resp := protos.GetTokenResponse{
			PublicAddress: param.PublicAddress,
			Token:         token,
		}
		utils.Response(ctx, utils.SuccessCode, utils.Success, resp)
	}
}

func (u *userApi) Register(ctx *gin.Context) {
	var param protos.User
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.InvalidParamErr.Message = err.Error()
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(param.PublicAddress) || !utils.IsValidAddress(param.PublicAddress) {
		utils.InvalidParamErr.Message = "Please enter correct pbulic_address."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(param.Email) || !utils.VerifyEmailFormat(param.Email) {
		utils.InvalidParamErr.Message = "Please enter correct email."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(param.Name) {
		utils.InvalidParamErr.Message = "Please enter name."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if len(param.Addresses) <= 0 {
		utils.InvalidParamErr.Message = "Please enter address."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}
	err := u.srv.CreateUser(ctx, &param)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	utils.Response(ctx, utils.SuccessCode, utils.Success, nil)
}

func (u *userApi) UpdateUser(ctx *gin.Context) {
	token, err := getToken(ctx)
	if err != nil {
		utils.InvalidParamErr.Message = "Please carry token."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	var param protos.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.InvalidParamErr.Message = err.Error()
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if utils.IsEmpty(param.PublicAddress) ||
		!utils.IsValidAddress(param.PublicAddress) ||
		token.PublicAddress != param.PublicAddress {
		utils.InvalidParamErr.Message = "Please enter correct address."
		utils.Response(ctx, utils.SuccessCode, utils.InvalidParamErr, nil)
		return
	}

	if len(param.UpdateMask) <= 0 {
		utils.InvalidParamErr.Message = "Please enter update mask."
		utils.Response(ctx, http.StatusOK, utils.InvalidParamErr, nil)
		return
	}
	userInfo, err := u.srv.UpdateUserInfo(ctx, token.PublicAddress, param.User, param.UpdateMask)
	if err != nil {
		utils.InternalServerError.Message = fmt.Sprintf("Operation failed, %s.", err.Error())
		utils.Response(ctx, utils.SuccessCode, utils.InternalServerError, nil)
		return
	}
	userInfo.PublicAddress = token.PublicAddress
	utils.Response(ctx, utils.SuccessCode, utils.Success, userInfo)
}

func getToken(ctx *gin.Context) (*protos.UserToken, error) {
	var userToken = new(protos.UserToken)
	if info, ok := ctx.Get("access_token"); !ok {
		return nil, errors.New("please carry token")
	} else {
		userToken = info.(*protos.UserToken)
	}
	return userToken, nil
}
