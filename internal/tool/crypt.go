package tool

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/argon2"
	"log"
	"math"
)

var (
	SaltByteLength   uint = 16
	SaltStringLength      = uint(math.Ceil(float64(SaltByteLength)/3)) * 4
)

func PasswordEncrypt(password string) string {
	salt := generateRandomByte(SaltByteLength)
	passwordByte := []byte(password)
	encryptedPassword := encrypt(passwordByte, salt)

	saltStr := base64.StdEncoding.EncodeToString(salt)
	hashedPasswordStr := base64.StdEncoding.EncodeToString(encryptedPassword)
	result := saltStr + hashedPasswordStr
	return result
}

func PasswordCompare(comparePassword, originalPassword string) bool {
	originalPasswordSalt, _ := saltStrDecode(originalPassword[0:SaltStringLength])
	comparePasswordByte := []byte(comparePassword)
	compareHashedPassword := encrypt(comparePasswordByte, originalPasswordSalt)
	if originalPassword == (originalPassword[0:SaltStringLength] + base64.StdEncoding.EncodeToString(compareHashedPassword)) {
		return true
	}
	return false
}

func saltStrDecode(saltStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(saltStr)
}

func encrypt(plaintext, salt []byte) []byte {
	return argon2.Key(plaintext, salt, 1, 32*1024, 4, 32)
}

func generateRandomByte(size uint) (data []byte) {
	data = make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal("")
	}
	return
}
