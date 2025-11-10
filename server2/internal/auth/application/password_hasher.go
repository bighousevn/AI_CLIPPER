package application

// PasswordHasher defines the interface for hashing and verifying passwords.
// This abstraction allows the use case to be independent of the specific hashing algorithm.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}