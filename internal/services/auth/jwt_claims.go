package auth

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

var jwtClaims = struct{}{}

type JWTClaims struct {
	jwt.RegisteredClaims

	UserID uuid.UUID `json:"uid"`
	Login  string    `json:"login"`
}

func FromContext(ctx context.Context) *JWTClaims {
	val := ctx.Value(jwtClaims)
	if val == nil {
		return nil
	}
	claims, ok := val.(*JWTClaims)
	if !ok {
		return nil
	}

	return claims
}

func ToContext(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, jwtClaims, claims)
}
