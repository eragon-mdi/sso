package tokeneradapter

import (
	"crypto/rsa"
	"os"

	"github.com/eragon-mdi/sso/internal/domain"
	authservice "github.com/eragon-mdi/sso/internal/service/sso/auth"
	"github.com/go-faster/errors"
	"github.com/golang-jwt/jwt/v5"

	"github.com/eragon-mdi/go-playground/tokener"
)

func New(privateRSAFilePath, publicRSAFilePath string) (authservice.Tokener, error) {
	priv, pub, err := loadKeys(privateRSAFilePath, publicRSAFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "load keys")
	}

	return &tokenerAdapter{
		Tokener: tokener.NewRSA(priv, pub),
	}, nil
}

type tokenerAdapter struct {
	tokener.Tokener
}

func (t *tokenerAdapter) GenPair(m domain.Meta) (access, refresh []byte, _ error) {
	return t.Tokener.GenPair(tokener.Claims(m), "access", "refresh")
}

type UnClaims interface {
	UnClaims(map[string]any) error
}

func (t *tokenerAdapter) VerifyRefresh(refresh []byte) (m domain.Meta, _ error) {
	claims, err := t.Tokener.VerifyRefresh(refresh)
	if err != nil {
		return domain.Meta{}, errors.Wrap(err, "failed verify")
	}

	if err := m.UnClaims(claims); err != nil {
		return domain.Meta{}, errors.Wrap(err, "failed unclaim")
	}

	return m, nil
}

func loadKeys(pathPrivate, pathPublic string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privBytes, err := os.ReadFile(pathPrivate)
	if err != nil {
		return nil, nil, errors.Wrap(err, "read private key")
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse private key")
	}

	pubBytes, err := os.ReadFile(pathPublic)
	if err != nil {
		return nil, nil, errors.Wrap(err, "read public key")
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse public key")
	}

	return privKey,
		pubKey, nil
}
