package erc20

import "errors"

var (
	// ErrInvalidAddress is returned when an invalid address is provided.
	ErrInvalidAddress = errors.New("invalid address")

	// ErrInvalidAmount is returned when an invalid amount is provided.
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrInvalidGasLimit is returned when an invalid gas limit is provided.
	ErrInvalidGasLimit = errors.New("invalid gas limit")

	// ErrInvalidGasFeeCap is returned when an invalid gas fee cap is provided.
	ErrInvalidGasFeeCap = errors.New("invalid gas fee cap")

	// ErrInvalidGasTipCap is returned when an invalid gas tip cap is provided.
	ErrInvalidGasTipCap = errors.New("invalid gas tip cap")

	// ErrInvalidGasPrice is returned when an invalid gas price is provided.
	ErrInvalidGasPrice = errors.New("invalid gas price")

	// ErrInvalidSignature is returned when an invalid signature is provided.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidNonce is returned when an invalid nonce is provided.
	ErrInvalidNonce = errors.New("invalid nonce")

	// ErrEthClient is returned when ethereum client error.
	ErrEthClient = errors.New("ethereum client error")

	// ErrSign is returned when sign error.
	ErrSign = errors.New("sign error")

	// ErrContractPack is returned when contract pack error.
	ErrContractPack = errors.New("contract pack error")

	// ErrContractUnpack is returned when contract unpack error.
	ErrContractUnpack = errors.New("contract unpack error")

	// ErrInvaildField is returned when invaild field.
	ErrInvaildField = errors.New("invaild field")
)
