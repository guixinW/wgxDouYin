package tool

import (
	"encoding/base64"
	"math/rand"
	"testing"
)

func TestPasswordEncrypt(t *testing.T) {
	randomKeyLength := 10 + rand.Intn(10)
	randomKey := RandomBytes(randomKeyLength)
	PasswordEncrypt(base64.StdEncoding.EncodeToString(randomKey))
}
