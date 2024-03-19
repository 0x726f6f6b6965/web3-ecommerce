package services

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/0x726f6f6b6965/web3-ecommerce/internal/helper"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang-jwt/jwt/v5"
)

func TestGetToken(t *testing.T) {
	os.Setenv("PRIVATE_KEY", "./private_key")
	os.Setenv("RPC", "wss://ethereum-sepolia-rpc.publicnode.com")
	os.Setenv("JWT_SECRET", "secret")
	helper.JwtSecretKey = []byte(os.Getenv("JWT_SECRET"))
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

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		t.Fatal(errors.Join(errors.New("get nonce error"), err))
	}

	msg := fmt.Sprintf(LoginMsg, from.Hex(), nonce)
	data := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg))
	hash := crypto.Keccak256Hash(data)

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		t.Fatal(errors.Join(errors.New("signature error"), err))
	}
	signature[64] += 27

	tokenString, err := NewUserService(client).GetToken(ctx, from.Hex(), hexutil.Encode(signature))
	if err != nil {
		t.Fatal(errors.Join(errors.New("get token error"), err))
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return helper.JwtSecretKey, nil
	})
	if err != nil {
		t.Fatal(errors.Join(errors.New("parse token error"), err))
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("error casting token claims to map claims")
	}
	if claims["user"] != from.Hex() {
		t.Fatal("user not match")
	}
}
