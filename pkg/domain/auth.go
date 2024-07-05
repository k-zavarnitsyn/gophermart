package domain

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/k-zavarnitsyn/gophermart/internal/config"
)

type hasher struct {
	cfg *config.Auth
}

func (h *hasher) GenerateSHA(password []byte) []byte {
	hs256 := hmac.New(sha256.New, h.cfg.PasswordHashKey)
	hs256.Write(password)

	return hs256.Sum(nil)
}
