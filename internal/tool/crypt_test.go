package tool

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestPasswordCompare(t *testing.T) {
	randomKeyLength := uint(10 + rand.Intn(10))
	randomKey1 := generateRandomByte(randomKeyLength)
	randomKey1Str := base64.StdEncoding.EncodeToString(randomKey1)
	cipText1 := PasswordEncrypt(randomKey1Str)
	start := time.Now()
	isEqual := PasswordCompare(randomKey1Str, cipText1)
	end := time.Now().Sub(start)
	fmt.Printf("密钥比较耗时:%v\n", end)
	if !isEqual {
		t.Fatalf("相同的明文与盐值加密为了不同的密文")
	}
}
