package hasher

import "golang.org/x/crypto/bcrypt"

type Hasher interface {
	Hash(password string) (string, error)
	Compare(password string, hash string) error
}

type BCryptHasher struct{}

var DefaultHasher = BCryptHasher{}

func (BCryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}

func (BCryptHasher) Compare(raw string, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw))
}
