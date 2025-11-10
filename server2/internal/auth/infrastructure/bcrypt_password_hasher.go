package infrastructure

import "golang.org/x/crypto/bcrypt"

// BcryptPasswordHasher is a concrete implementation of the PasswordHasher interface using bcrypt.
type BcryptPasswordHasher struct{}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{}
}

// Hash takes a plain password and returns its bcrypt hash.
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Check compares a plain password with a hash to see if they match.
func (h *BcryptPasswordHasher) Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}