package tool

import (
	"encoding/base64"
	"math/rand"
	"testing"
)

func TestPasswordCompare(t *testing.T) {
	randomKeyLength := uint(10 + rand.Intn(10))
	randomKey1 := generateRandomByte(randomKeyLength)
	randomKey1Str := base64.StdEncoding.EncodeToString(randomKey1)

	cipText1 := PasswordEncrypt(randomKey1Str)
	if !PasswordCompare(randomKey1Str, cipText1) {
		t.Fatalf("相同的明文与盐值加密为了不同的密文")
	}
}
