package services

import "errors"

var (
	ErrDynamodbClientNotFound = errors.New("dynamodb client not found")
	ErrAlreadyPaid            = errors.New("already paid")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrDynamodb               = errors.New("dynamodb operation failed")
	ErrInvalidSignature       = errors.New("invalid signature")
	ErrGenerateToken          = errors.New("generate token failed")
)
