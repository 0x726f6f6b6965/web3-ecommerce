package erc20

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/contract"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	TRANSFER   = "transfer"
	BALANCE_OF = "balanceOf"
	APPROVE    = "approve"
	ALLOWANCE  = "allowance"

	EVENT_TRANSFER = "Transfer"
)

type ERC20Service interface {
	// TransferWithPrivateKey - send a transaction with private key.
	// @param ctx - context
	// @param trans - common request
	// @param privateKey - private key
	// @return transaction
	// @return error
	TransferWithPrivateKey(ctx context.Context, trans protos.CommonRequest, privateKey *ecdsa.PrivateKey) (*types.Transaction, error)
	// TransferWithSign - send a transaction with sign.
	// @param ctx - context
	// @param trans - common request
	// @return transaction
	// @return error
	TransferWithSign(ctx context.Context, trans protos.CommonRequest) (*types.Transaction, error)
	// BalanceOf - get balance of an address.
	// @param ctx - context
	// @param address - address
	// @return balance
	// @return error
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
	// SubscribeTransfer - subscribe transfer event.
	// @param process - process function
	// @param fromBlock - from block
	// @return unsubscribe function
	// @return error channel
	SubscribeTransfer(process func(types.Log), fromBlock *big.Int) (func(), <-chan error)
	// SubscribeTransferWithFilter - subscribe transfer event with filter.
	// @param process - process function
	// @param fromBlock - from block
	// @param filter - filter
	// @return unsubscribe function
	// @return error channel
	SubscribeTransferWithFilter(process func(types.Log), fromBlock *big.Int, filter ethereum.FilterQuery) (func(), <-chan error)
	// ApproveWithPrivateKey - approve an address to transfer token with private key.
	// @param ctx - context
	// @param request - common request
	// @param privateKey - private key
	// @return transaction
	// @return error
	ApproveWithPrivateKey(ctx context.Context, request protos.CommonRequest, privateKey *ecdsa.PrivateKey) (*types.Transaction, error)
	// ApproveWithSign - approve an address to transfer token with sign.
	// @param ctx - context
	// @param request - common request
	// @return transaction
	// @return error
	ApproveWithSign(ctx context.Context, request protos.CommonRequest) (*types.Transaction, error)
	// CheckAllowance - check allowance of an address.
	// @param ctx - context
	// @param request - common request
	// @return allowance
	// @return error
	CheckAllowance(ctx context.Context, request protos.CheckAllowanceRequest) (*big.Int, error)
	// GetABI - get contract of abi
	// @return abi
	GetABI() abi.ABI
}

type service struct {
	client   *ethclient.Client
	contract *contract.Contract
	chainId  *big.Int
	decimals int
}

func NewERC20Service(client *ethclient.Client, contract *contract.Contract, chainId *big.Int, decimals int) ERC20Service {

	return &service{
		client:   client,
		contract: contract,
		chainId:  chainId,
		decimals: decimals,
	}
}

func (s *service) TransferWithPrivateKey(ctx context.Context, trans protos.CommonRequest, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	input, err := s.checkCommonRequest(trans, TRANSFER)
	if err != nil {
		return nil, err
	}
	if trans.Amount == 0 {
		return nil, errors.Join(ErrInvalidAmount, fmt.Errorf("amount field is 0"))
	}

	from := common.HexToAddress(trans.From)

	params := new(ethereum.CallMsg)
	params.From = from
	params.To = &s.contract.Address
	params.Data = input

	var gas uint64
	if utils.IsEmpty(trans.Gas) {
		gas, err = s.client.EstimateGas(ctx, *params)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
	} else {
		val := big.NewInt(0)
		val, ok := val.SetString(trans.Gas, 10)
		if !ok {
			return nil, ErrInvalidGasLimit
		}
		gas = val.Uint64()
	}
	params.Gas = gas

	var gasPrice *big.Int
	if utils.IsEmpty(trans.GasFeeCap) {
		gasPrice, err = s.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
	} else {
		val := big.NewInt(0)
		val, ok := val.SetString(trans.GasFeeCap, 10)
		if !ok {
			return nil, ErrInvalidGasFeeCap
		}
		gasPrice = val
	}
	params.GasFeeCap = gasPrice

	var gasTipCap *big.Int
	if utils.IsEmpty(trans.GasTipCap) {
		gasTipCap, err = s.client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
	} else {
		val := big.NewInt(0)
		val, ok := val.SetString(trans.GasTipCap, 10)
		if !ok {
			return nil, ErrInvalidGasTipCap
		}
		gasTipCap = val
	}
	params.GasTipCap = gasTipCap

	return s.transaction(ctx, trans.Nonce, *params, nil, privateKey)
}

func (s *service) TransferWithSign(ctx context.Context, trans protos.CommonRequest) (*types.Transaction, error) {
	input, err := s.checkCommonRequest(trans, TRANSFER)
	if err != nil {
		return nil, err
	}
	if trans.Amount == 0 {
		return nil, errors.Join(ErrInvalidAmount, fmt.Errorf("amount field is 0"))
	}

	if utils.IsEmpty(trans.Signature) {
		return nil, errors.Join(ErrInvalidSignature, fmt.Errorf("signature field is empty"))
	}

	from := common.HexToAddress(trans.From)
	params := new(ethereum.CallMsg)
	params.From = from
	params.To = &s.contract.Address
	params.Data = input

	if utils.IsEmpty(trans.GasTipCap) {
		return nil, errors.Join(ErrInvalidGasTipCap, fmt.Errorf("gasTipCap field is empty"))
	}
	gasTip := big.NewInt(0)
	gasTip, ok := gasTip.SetString(trans.GasTipCap, 10)
	if !ok {
		return nil, ErrInvalidGasTipCap
	}
	params.GasTipCap = gasTip

	if utils.IsEmpty(trans.Gas) {
		return nil, errors.Join(ErrInvalidGasPrice, fmt.Errorf("gas field is empty"))
	}
	gas := big.NewInt(0)
	gas, ok = gas.SetString(trans.Gas, 10)
	if !ok {
		return nil, ErrInvalidGasPrice
	}
	params.Gas = gas.Uint64()

	if utils.IsEmpty(trans.GasFeeCap) {
		return nil, errors.Join(ErrInvalidGasPrice, fmt.Errorf("gasFeeCap field is empty"))
	}
	gasFee := big.NewInt(0)
	gasFee, ok = gasFee.SetString(trans.GasFeeCap, 10)
	if !ok {
		return nil, ErrInvalidGasFeeCap
	}
	params.GasFeeCap = gasFee
	var signature []byte
	if strings.HasPrefix(trans.Signature, "0x") {
		signature, err = hexutil.Decode(trans.Signature)
		if err != nil {
			return nil, errors.Join(ErrInvalidSignature, err)
		}
	} else {
		signature = []byte(trans.Signature)
	}

	return s.transaction(ctx, trans.Nonce, *params, signature, nil)
}

func (s *service) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	if utils.IsEmpty(address) || utils.IsValidAddress(address) {
		return nil, ErrInvalidAddress
	}
	addr := common.HexToAddress(address)
	input, err := s.contract.ABI.Pack(BALANCE_OF, addr)
	if err != nil {
		return nil, errors.Join(ErrContractPack, err)
	}
	callParams := ethereum.CallMsg{
		From: addr,
		To:   &s.contract.Address,
		Data: input,
	}
	data, err := s.client.CallContract(ctx, callParams, nil)
	if err != nil {
		return nil, errors.Join(ErrEthClient, err)
	}
	outputs := make(map[string]interface{})
	err = s.contract.ABI.UnpackIntoMap(outputs, BALANCE_OF, data)
	if err != nil {
		return nil, errors.Join(ErrContractUnpack, err)
	}

	if balance, ok := outputs["balance"].(*big.Int); ok {
		return balance, nil
	}
	return nil, errors.Join(ErrContractUnpack, fmt.Errorf("field balance not found"))
}

func (s *service) SubscribeTransfer(process func(types.Log), fromBlock *big.Int) (func(), <-chan error) {
	if fromBlock == nil {
		return nil, nil
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{s.contract.Address},
		Topics: [][]common.Hash{
			{s.contract.ABI.Events[EVENT_TRANSFER].ID}},
		FromBlock: fromBlock,
	}

	logs := make(chan types.Log)
	errChan := make(chan error)
	var stop chan struct{}
	go func() {
		sub, err := s.client.SubscribeFilterLogs(context.Background(), query, logs)
		if err != nil {
			errChan <- err
			return
		}

		for {
			var t <-chan time.Time
			if stop == nil {
				t = time.After(time.Second * 3)
			}
			select {
			case <-stop:
				sub.Unsubscribe()
				close(errChan)
				close(stop)
				return
			case err := <-sub.Err():
				errChan <- err
				sub.Unsubscribe()
				return
			case vLog := <-logs:
				process(vLog)
			case <-t:
				continue
			}
		}
	}()
	return func() {
		stop = make(chan struct{})
		stop <- struct{}{}
	}, errChan
}

func (s *service) SubscribeTransferWithFilter(process func(types.Log), fromBlock *big.Int, filter ethereum.FilterQuery) (func(), <-chan error) {
	if fromBlock == nil {
		return nil, nil
	}
	filter.Addresses = []common.Address{s.contract.Address}
	logs := make(chan types.Log)
	errChan := make(chan error)
	var stop chan struct{}
	go func() {
		sub, err := s.client.SubscribeFilterLogs(context.Background(), filter, logs)
		if err != nil {
			errChan <- err
			return
		}

		for {
			var t <-chan time.Time
			if stop == nil {
				t = time.After(time.Second * 3)
			}
			select {
			case <-stop:
				sub.Unsubscribe()
				close(errChan)
				close(stop)
				return
			case err := <-sub.Err():
				errChan <- err
				sub.Unsubscribe()
				return
			case vLog := <-logs:
				process(vLog)
			case <-t:
				continue
			}
		}
	}()
	return func() {
		stop = make(chan struct{})
		stop <- struct{}{}
	}, errChan
}

func (s *service) ApproveWithPrivateKey(ctx context.Context, request protos.CommonRequest, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	input, err := s.checkCommonRequest(request, APPROVE)
	if err != nil {
		return nil, err
	}
	from := common.HexToAddress(request.From)

	params := new(ethereum.CallMsg)
	params.From = from
	params.To = &s.contract.Address
	params.Data = input

	var gas uint64
	if utils.IsEmpty(request.Gas) {
		gas, err = s.client.EstimateGas(ctx, *params)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
	} else {
		val := big.NewInt(0)
		val, ok := val.SetString(request.Gas, 10)
		if !ok {
			return nil, ErrInvalidGasLimit
		}
		gas = val.Uint64()
	}
	params.Gas = gas

	var gasPrice *big.Int
	if utils.IsEmpty(request.GasFeeCap) {
		gasPrice, err = s.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
	} else {
		val := big.NewInt(0)
		val, ok := val.SetString(request.GasFeeCap, 10)
		if !ok {
			return nil, ErrInvalidGasFeeCap
		}
		gasPrice = val
	}
	params.GasFeeCap = gasPrice

	var gasTipCap *big.Int
	if utils.IsEmpty(request.GasTipCap) {
		gasTipCap, err = s.client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
	} else {
		val := big.NewInt(0)
		val, ok := val.SetString(request.GasTipCap, 10)
		if !ok {
			return nil, ErrInvalidGasTipCap
		}
		gasTipCap = val
	}
	params.GasTipCap = gasTipCap

	return s.transaction(ctx, request.Nonce, *params, nil, privateKey)
}

func (s *service) ApproveWithSign(ctx context.Context, request protos.CommonRequest) (*types.Transaction, error) {
	input, err := s.checkCommonRequest(request, APPROVE)
	if err != nil {
		return nil, err
	}

	if utils.IsEmpty(request.Signature) {
		return nil, errors.Join(ErrInvalidSignature, fmt.Errorf("signature field is empty"))
	}

	params := new(ethereum.CallMsg)

	from := common.HexToAddress(request.From)
	params.From = from
	params.Data = input
	params.To = &s.contract.Address

	if utils.IsEmpty(request.GasTipCap) {
		return nil, errors.Join(ErrInvalidGasTipCap, fmt.Errorf("gasTipCap field is empty"))
	}
	gasTip := big.NewInt(0)
	gasTip, ok := gasTip.SetString(request.GasTipCap, 10)
	if !ok {
		return nil, ErrInvalidGasTipCap
	}
	params.GasTipCap = gasTip

	if utils.IsEmpty(request.Gas) {
		return nil, errors.Join(ErrInvalidGasPrice, fmt.Errorf("gas field is empty"))
	}
	gas := big.NewInt(0)
	gas, ok = gas.SetString(request.Gas, 10)
	if !ok {
		return nil, ErrInvalidGasPrice
	}
	params.Gas = gas.Uint64()

	if utils.IsEmpty(request.GasFeeCap) {
		return nil, errors.Join(ErrInvalidGasPrice, fmt.Errorf("gasFeeCap field is empty"))
	}
	gasFee := big.NewInt(0)
	gasFee, ok = gasFee.SetString(request.GasFeeCap, 10)
	if !ok {
		return nil, ErrInvalidGasFeeCap
	}
	params.GasFeeCap = gasFee

	return s.transaction(ctx, request.Nonce, *params, []byte(request.Signature), nil)
}

func (s *service) CheckAllowance(ctx context.Context, request protos.CheckAllowanceRequest) (*big.Int, error) {
	if utils.IsEmpty(request.From) || utils.IsValidAddress(request.From) {
		return nil, errors.Join(ErrInvalidAddress, fmt.Errorf("from address is invalid"))
	}

	if utils.IsEmpty(request.To) || utils.IsValidAddress(request.To) {
		return nil, errors.Join(ErrInvalidAddress, fmt.Errorf("to address is invalid"))
	}
	from := common.HexToAddress(request.From)
	to := common.HexToAddress(request.To)
	input, err := s.contract.ABI.Pack(ALLOWANCE, from, to)
	if err != nil {
		return nil, errors.Join(ErrContractPack, err)
	}
	callParams := ethereum.CallMsg{
		From: from,
		To:   &s.contract.Address,
		Data: input,
	}
	data, err := s.client.CallContract(ctx, callParams, nil)
	if err != nil {
		return nil, errors.Join(ErrEthClient, err)
	}
	var output interface{}
	err = s.contract.ABI.UnpackIntoInterface(output, ALLOWANCE, data)
	if err != nil {
		return nil, errors.Join(ErrContractUnpack, err)
	}

	if num, ok := output.(*big.Int); ok {
		return num, nil
	}
	return nil, nil
}

func (s *service) GetABI() abi.ABI {
	return s.contract.ABI
}

func (s *service) checkCommonRequest(request protos.CommonRequest, method string) ([]byte, error) {
	if utils.IsEmpty(request.From) || !utils.IsValidAddress(request.From) {
		return nil, errors.Join(ErrInvalidAddress, fmt.Errorf("from address is invalid"))
	}

	if utils.IsEmpty(request.To) || !utils.IsValidAddress(request.To) {
		return nil, errors.Join(ErrInvalidAddress, fmt.Errorf("to address is invalid"))
	}
	to := common.HexToAddress(request.To)

	if request.Amount < 0 {
		return nil, ErrInvalidAmount
	}
	amount := contract.ToWei(request.Amount, 6)

	input, err := s.contract.ABI.Pack(method, to, amount)
	if err != nil {
		return nil, errors.Join(ErrContractPack, err)
	}
	return input, nil
}

func (s *service) transaction(ctx context.Context, nonce uint64, callParams ethereum.CallMsg, sign []byte, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: callParams.GasTipCap,
		GasFeeCap: callParams.GasFeeCap,
		Gas:       callParams.Gas,
		To:        callParams.To,
		Data:      callParams.Data,
		Value:     callParams.Value,
	})

	signer := types.NewLondonSigner(s.chainId)

	if privateKey != nil {
		signTx, err := types.SignTx(tx, signer, privateKey)
		if err != nil {
			return nil, errors.Join(ErrSign, err)
		}
		err = s.client.SendTransaction(ctx, signTx)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
		return signTx, nil
	}

	if sign != nil {
		signTx, err := tx.WithSignature(signer, sign)
		if err != nil {
			return nil, errors.Join(ErrSign, err)
		}

		sigPublicKey, err := signer.Sender(signTx)
		if err != nil {
			return nil, errors.Join(ErrSign, err)
		}
		if matches := bytes.Equal(sigPublicKey.Bytes(), callParams.From.Bytes()); !matches {
			return nil, ErrInvalidSignature
		}

		err = s.client.SendTransaction(ctx, signTx)
		if err != nil {
			return nil, errors.Join(ErrEthClient, err)
		}
		return signTx, nil
	}
	return nil, ErrInvalidSignature
}
