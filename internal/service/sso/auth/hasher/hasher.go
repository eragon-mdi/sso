package hasher

import (
	authservice "github.com/eragon-mdi/sso/internal/service/sso/auth"
	"golang.org/x/crypto/bcrypt"
)

type hasher struct {
	cost int
}

func New(cost int) authservice.PasswordHasher {
	return &hasher{
		cost: cost,
	}
}

func (h *hasher) Gen(origin []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(origin), h.cost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (h *hasher) Compare(hash []byte, origin []byte) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(hash, origin); err != nil {
		return false, err
	}

	return true, nil
}
