package util

import (
	"golang.org/x/crypto/bcrypt"
	"wgxDouYin/internal/user/domain/service"
)

// BcryptPasswordHasher provides a bcrypt-based implementation of the PasswordService.
type BcryptPasswordHasher struct{}

// NewBcryptPasswordHasher creates a new instance of BcryptPasswordHasher.
func NewBcryptPasswordHasher() service.PasswordService {
	return &BcryptPasswordHasher{}
}

// Hash generates a bcrypt hash from a password.
func (b *BcryptPasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Check compares a hashed password with a plaintext password.
func (b *BcryptPasswordHasher) Check(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
