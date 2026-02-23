package service

import "golang.org/x/crypto/bcrypt"

const (
	DefaultHashCost = 14
	MinHashCost     = 12
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < MinHashCost {
		cost = MinHashCost
	}

	return &BcryptHasher{
		cost: cost,
	}
}

func (s *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	return string(bytes), err
}

func (s *BcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
