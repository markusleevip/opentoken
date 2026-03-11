package utils

import (
	"crypto/sha1"
	"encoding/hex"
)

func Sha1Hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
