package utils

const (
	SuccessCode                    = 200
	ErrorCode                      = -1
	ErrorCodeOfInternalServerError = 100        // internal server error, please check server log
	ErrorCodeOfInvalidParams       = 100 + iota // param error
	ProtocolClientErrCode
	ProtocolAssetTypeErrCode

	ErrorCodeLogin    = 401
	ErrorCodeNotFound = 404
)

var (
	Success              = ErrorString{SuccessCode, "success"}
	InvalidParamErr      = ErrorString{ErrorCodeOfInvalidParams, "Wrong request parameter"}
	InternalServerError  = ErrorString{ErrorCodeOfInternalServerError, "Service internal exception"}
	ErrorCodeLoginError  = ErrorString{ErrorCodeLogin, "The account is not logged in, please login and try again"}
	ErrorCodeNotFoundErr = ErrorString{ErrorCodeNotFound, "The resource is not found"}

	ProtocolClientErr    = ErrorString{Code: ProtocolClientErrCode, Message: "client id or client secret error"}
	ProtocolAssetTypeErr = ErrorString{Code: ProtocolAssetTypeErrCode, Message: "asset type invalid"}
)

type ErrorString struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
