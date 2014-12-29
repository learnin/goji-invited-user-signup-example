package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func Hash(s string, salt string) string {
	hash := sha256.New()
	io.WriteString(hash, s+salt)
	return hex.EncodeToString(hash.Sum(nil))
}
