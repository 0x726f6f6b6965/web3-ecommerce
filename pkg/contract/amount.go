package contract

import (
	"math/big"

	"github.com/shopspring/decimal"
)

func ToWei(val float64, decimals int) *big.Int {

	amount := decimal.NewFromFloat(val)
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)
	return wei
}

func ToAmount(val string, decimals int) (*big.Int, error) {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result, err := decimal.NewFromString(val)
	if err != nil {
		return nil, err
	}
	return result.Div(mul).BigInt(), nil
}
