package main

import "errors"

var (
	ErrInvalidEvent error = errors.New("invalid event")
	ErrUnmarshal    error = errors.New("unmarshal error")
	ErrTimeout      error = errors.New("timeout")
	ErrMonitor      error = errors.New("monitor error")
	ErrUpdateTrans  error = errors.New("update transaction error")
)
