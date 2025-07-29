package service

// PasswordService provides an interface for password-related operations,
// abstracting the specific hashing algorithm.
type PasswordService interface {
	// Hash creates a hashed string from a plaintext password.
	Hash(password string) (string, error)
	// Check compares a plaintext password with a hashed password to see if they match.
	Check(hashedPassword, password string) bool
}
