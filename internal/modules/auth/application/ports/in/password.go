package in

// PasswordService defines the interface for password hashing and verification.
type PasswordService interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}
