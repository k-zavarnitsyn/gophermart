package middleware

import (
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
)

type Auth struct {
	authService *auth.Service
}

func NewAuth(authService *auth.Service) *Auth {
	return &Auth{
		authService: authService,
	}
}

func (a *Auth) WithAuthentication(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.authService.Authenticate(r)
		if err != nil {
			utils.SendErrorMsg(w, err, "unable to authenticate user", http.StatusUnauthorized)
			return
		}
		ctx := auth.ToContext(r.Context(), claims)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
