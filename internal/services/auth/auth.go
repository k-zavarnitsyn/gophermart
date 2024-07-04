package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

var (
	ErrInvalidToken               = errors.New("invalid token")
	ErrTokenNotProvided           = errors.New("no token provided")
	ErrUserIDNotProvided          = errors.New("no token user ID provided")
	ErrUserLoginNotProvided       = errors.New("no token user login provided")
	ErrTokenExpirationNotProvided = errors.New("no token expiration time provided")
)

type Service struct {
	cfg          *config.Auth
	jwtParser    *jwt.Parser
	jwtPublicKey any
}

func New(config *config.Auth) *Service {
	return &Service{
		cfg: config,
		jwtParser: jwt.NewParser(
			jwt.WithValidMethods(config.ValidMethods),
			jwt.WithLeeway(config.Leeway),
			jwt.WithExpirationRequired(),
		),
		jwtPublicKey: config.JwtPrivateKey.Public(),
	}
}

func (s *Service) Authenticate(r *http.Request) (*JWTClaims, error) {
	strToken, err := s.GetToken(r)
	if err != nil {
		return nil, err
	}

	return s.ParseToken(strToken)
}

func (s *Service) GetToken(r *http.Request) (string, error) {
	c, err := r.Cookie(s.cfg.CookieName)
	if err != nil {
		return "", err
	}

	return c.Value, nil
}

func (s *Service) ParseToken(token string) (*JWTClaims, error) {
	if token == "" {
		return nil, ErrTokenNotProvided
	}

	claims := &JWTClaims{}
	jwtToken, err := s.jwtParser.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtPublicKey, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) NewClaims(user *entity.User) *JWTClaims {
	return &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.ExpiresIn)),
		},
		UserID: user.ID,
		Login:  user.Login,
	}
}

func (s *Service) CreateToken(claims *JWTClaims) (string, error) {
	if claims.UserID.IsNil() {
		return "", ErrUserIDNotProvided
	}
	if claims.Login == "" {
		return "", ErrUserLoginNotProvided
	}
	if claims.ExpiresAt == nil {
		return "", ErrTokenExpirationNotProvided
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(s.cfg.JwtPrivateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Service) CreateTokenCookie(user *entity.User) (*http.Cookie, error) {
	claims := s.NewClaims(user)
	tokenString, err := s.CreateToken(claims)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:    s.cfg.CookieName,
		Value:   tokenString,
		Expires: claims.ExpiresAt.Time,
	}, nil
}
