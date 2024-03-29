package utils

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var ErrInvalidSig = errors.New("invalid transaction v, r, s values")

func GetSignFromDynamicFeeTxType(R, S, Vb *big.Int, homestead bool) ([]byte, error) {
	Vb = new(big.Int).Add(Vb, big.NewInt(27))
	if Vb.BitLen() > 8 {
		return nil, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return nil, ErrInvalidSig
	}
	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	return sig, nil
}

// VerifySignature checks the signature of the given message.
func VerifySignature(from, sigHex, msg string) error {
	// input validation
	sig, err := hexutil.Decode(sigHex)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %v", err.Error())
	}
	if len(sig) != 65 {
		return fmt.Errorf("invalid Ethereum signature length: %v", len(sig))
	}
	if sig[crypto.RecoveryIDOffset] != 27 && sig[crypto.RecoveryIDOffset] != 28 {
		return fmt.Errorf("invalid Ethereum signature (V is not 27 or 28): %v", sig[64])
	}

	// calculate message hash
	msgHash := accounts.TextHash([]byte(msg))

	// recover public key from signature and verify it matches the from address
	sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	recovered, err := crypto.SigToPub(msgHash, sig)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %v", err.Error())
	}
	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if strings.EqualFold(from, recoveredAddr.Hex()) {
		return nil
	}
	return fmt.Errorf("invalid Ethereum signature (addresses don't match)")
}
