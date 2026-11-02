package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateOTP
func GenerateOTP() string {
	const otpLength = 6
	const allDigits   = "0123456789"
	const nonZeroDigits = "123456789"

	otp := make([]byte, otpLength)

	firstNum, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nonZeroDigits))))
	otp[0] = nonZeroDigits[firstNum.Int64()]

	for i := 1; i < otpLength; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allDigits))))
		otp[i] = allDigits[num.Int64()]
	}

	return string(otp)
}