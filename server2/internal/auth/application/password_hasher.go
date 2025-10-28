package application

// PasswordHasher defines the contract for a password hashing service.
// This allows the use case to be independent of the specific hashing algorithm.
type PasswordHasher interface {
	Hash(password string) (string, error)
	CheckPassword(password, hash string) bool
}
