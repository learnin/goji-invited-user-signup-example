package helpers

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func Hash(s string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(s + salt))
	return hex.EncodeToString(hash.Sum(nil))
}

func SSHA(s string, salt string) string {
	hash := sha1.New()
	hash.Write([]byte(s + salt))
	return "{SSHA}" + base64.StdEncoding.EncodeToString(append(hash.Sum(nil), []byte(salt)...))
}
