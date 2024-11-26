package tool

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"log"
	"math"
)

var (
	SaltByteLength   uint = 16
	SaltStringLength uint = uint(math.Ceil(float64(SaltByteLength)/3)) * 4
)

func RandomBytes(size uint) (data []byte) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatal("")
	}
	return salt
}

func PasswordEncrypt(password string) string {
	salt := RandomBytes(SaltByteLength)
	passwordByte := []byte(password)
	hashedPassword := argon2.Key(passwordByte, salt, 1, 32*1024, 4, 32)
	fmt.Println(hashedPassword)
	fmt.Println(salt)
	result := base64.StdEncoding.EncodeToString(salt) + base64.StdEncoding.EncodeToString(hashedPassword)
	fmt.Println(len(base64.StdEncoding.EncodeToString(salt)))
	fmt.Println(SaltStringLength)
	return result
}
