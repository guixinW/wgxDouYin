package tool

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateHashOfLength64(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}
