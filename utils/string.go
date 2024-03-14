package utils

import (
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Check string is empty
func IsEmpty(s string) bool {
	return strings.Trim(s, " ") == ""
}

// IsValidAddress validate hex address
func IsValidAddress[addr string | common.Address](val addr) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := any(val).(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// Converts a string to CamelCase
func ToCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := true
	prevIsCap := false
	for i, v := range []byte(s) {
		isCap := v >= 'A' && v <= 'Z'
		isLow := v >= 'a' && v <= 'z'

		if capNext || i == 0 {
			if isLow {
				v += 'A'
				v -= 'a'
			}
		} else if prevIsCap && isCap {
			v += 'a'
			v -= 'A'
		}

		prevIsCap = isCap

		if isCap || isLow {
			n.WriteByte(v)
			capNext = false
		} else if isNum := v >= '0' && v <= '9'; isNum {
			n.WriteByte(v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}

// VerifyEmailFormat email verify
func VerifyEmailFormat(email string) bool {
	pattern := `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}
