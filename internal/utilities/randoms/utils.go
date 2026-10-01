package randoms_utils

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"math/big"
	math_rand "math/rand"
)

func GenRandNumBetween(min, max int64) int64 {
	return math_rand.Int63n(max-min+1) + min
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func SecureRandString(min, max int) (string, error) {
	// choose secure random length
	diff := max - min + 1
	nBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return "", err
	}
	length := int(nBig.Int64()) + min

	// generate securely random bytes
	b := make([]byte, length)
	for i := range b {
		idxBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idxBig.Int64()]
	}

	return string(b), nil
}

func SecureRandStringWithPool(pool string, min, max int) (string, error) {
	// choose secure random length
	diff := max - min + 1
	nBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return "", err
	}
	length := int(nBig.Int64()) + min

	// generate securely random bytes
	b := make([]byte, length)
	for i := range b {
		idxBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(len(pool))))
		if err != nil {
			return "", err
		}
		b[i] = pool[idxBig.Int64()]
	}

	return string(b), nil
}

func GenerateResetToken() (string, error) {
	return generateCryptoRandValue(32)
}

func GenerateAuthidValue() (string, error) {
	return generateCryptoRandValue(32)
}

func GeneratePidValue() (string, error) {
	return generateCryptoRandValue(12)
}

func GenerateFeedbackSlugValue() (string, error) {
	return generateCryptoRandValue(64)
}

func GenerateImageFilename() (string, error) {
	return generateCryptoRandValue(32)
}

func GenerateConfirmToken() (string, error) {
	return generateCryptoRandValue(40)
}

func generateCryptoRandValue(l int) (string, error) {
	b := make([]byte, l)
	if _, err := crypto_rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
