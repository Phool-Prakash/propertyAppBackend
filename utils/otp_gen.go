package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOtp(phoneNumber string) (string, error) { // <-- Corrected function name
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random number: %w", err)
	}
	otpInt := n.Int64() + 100000

	return fmt.Sprintf("%06d", otpInt), nil
}