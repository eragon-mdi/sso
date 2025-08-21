package hashertokener

import (
	"crypto/hmac"
	"crypto/sha256"

	authservice "github.com/eragon-mdi/sso/internal/service/sso/auth"
	"github.com/go-faster/errors"
)

type Hasher struct {
	secret []byte
}

func New(secret []byte) authservice.TokenHasher {
	return &Hasher{
		secret: secret,
	}
}

func (h *Hasher) Sum(data []byte) ([]byte, error) {
	if len(h.secret) == 0 {
		return nil, errors.New("secret is empty")
	}

	mac := hmac.New(sha256.New, h.secret)
	_, err := mac.Write(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write data to HMAC")
	}

	return mac.Sum(nil), nil
}
