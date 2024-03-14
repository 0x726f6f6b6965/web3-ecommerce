package contract

import (
	"bytes"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Contract struct {
	ABI     abi.ABI
	Address common.Address
}

func CreateContract(abiFile string, contractKey string) (*Contract, error) {

	conetnt, err := os.ReadFile(os.Getenv(abiFile))
	if err != nil {
		return nil, err
	}

	targetABI, err := abi.JSON(bytes.NewReader(conetnt))
	if err != nil {
		return nil, err
	}

	contract := &Contract{
		ABI:     targetABI,
		Address: common.HexToAddress(contractKey),
	}
	return contract, nil
}
