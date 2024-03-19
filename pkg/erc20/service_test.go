package erc20

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/0x726f6f6b6965/web3-ecommerce/pkg/contract"
	"github.com/0x726f6f6b6965/web3-ecommerce/protos"
	"github.com/0x726f6f6b6965/web3-ecommerce/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestTransferWithSign(t *testing.T) {
	os.Setenv("ERC20", "./../../deployment/abi/erc-20.json")
	os.Setenv("PRIVATE_KEY", "./private_key")
	os.Setenv("RPC", "wss://ethereum-sepolia-rpc.publicnode.com")
	os.Setenv("TO", "./to")
	os.Setenv("USDC", "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238")

	pk, err := os.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		t.Fatal(errors.Join(errors.New("private key not found"), err))
	}
	privateKey, err := crypto.HexToECDSA(string(pk))
	if err != nil {
		t.Fatal(errors.Join(errors.New("invalid private key"), err))
	}
	client, err := ethclient.Dial(os.Getenv("RPC"))
	if err != nil {
		t.Fatal(errors.Join(errors.New("connect ethereum error"), err))
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("error casting public key to ECDSA")
	}
	from := crypto.PubkeyToAddress(*publicKeyECDSA)
	ctx := context.Background()
	toByte, err := os.ReadFile(os.Getenv("TO"))
	if err != nil {
		t.Fatal(errors.Join(errors.New("to address not found"), err))
	}
	to := common.BytesToAddress(toByte)
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get nonce error"), err))
	}

	req := protos.CommonRequest{
		From:   from.Hex(),
		To:     to.Hex(),
		Amount: 0.1,
		Nonce:  nonce,
	}

	usdc, err := contract.CreateContract("ERC20", os.Getenv("USDC"))
	if err != nil {
		t.Fatal(errors.Join(errors.New("create abi error"), err))
	}

	amount := contract.ToWei(req.Amount, 6)
	input, err := usdc.ABI.Pack(TRANSFER, to, amount)
	if err != nil {
		t.Fatal(errors.Join(ErrContractPack, err))
	}
	token := common.HexToAddress(os.Getenv("USDC"))
	params := new(ethereum.CallMsg)
	params.From = from
	params.To = &token
	params.Data = input
	gas, err := client.EstimateGas(ctx, *params)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get gas error"), err))
	}
	params.Gas = gas
	req.Gas = fmt.Sprintf("%d", gas)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get gas price error"), err))
	}
	params.GasFeeCap = gasPrice
	req.GasFeeCap = fmt.Sprintf("%d", gasPrice)
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get gas tip error"), err))
	}
	params.GasTipCap = gasTipCap
	req.GasTipCap = fmt.Sprintf("%d", gasTipCap)
	chain, err := client.ChainID(ctx)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get chain id error"), err))
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chain,
		Nonce:     nonce,
		GasTipCap: params.GasTipCap,
		GasFeeCap: params.GasFeeCap,
		Gas:       params.Gas,
		To:        params.To,
		Data:      params.Data,
		Value:     params.Value,
	})
	signer := types.NewLondonSigner(chain)
	signTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		t.Fatal(errors.Join(errors.New("signature error"), err))
	}

	v, r, s := signTx.RawSignatureValues()
	msg, err := utils.GetSignFromDynamicFeeTxType(r, s, v, true)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get raw signature error"), err))
	}
	req.Signature = hexutil.Encode(msg)
	_, err = NewERC20Service(client, usdc, chain, 6).TransferWithSign(ctx, req)
	if err != nil {
		t.Fatal(errors.Join(errors.New("TransferWithSign error"), err))
	}
}
