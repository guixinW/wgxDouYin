package tool

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/argon2"
	"log"
)

func RandomBytes(size int) (data []byte) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		log.Fatal("")
	}
	return salt
}

func Argon2Encrypt(password []byte, salt []byte) string {
	hashedPassword := argon2.Key(password, salt, 1, 32*1024, 4, 32)
	result := base64.StdEncoding.EncodeToString(salt) + "$" + base64.StdEncoding.EncodeToString(hashedPassword)
	return result
}

func PasswordEncrypt(password string) string {
	salt := RandomBytes(16)
	return Argon2Encrypt([]byte(password), salt)
}
