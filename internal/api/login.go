package api

import (
	"errors"
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

func (s *gophermartServer) Login(w http.ResponseWriter, r *http.Request) {
	loginReq, err := utils.ReadJSON[entity.LoginRequest](r.Body)
	if err != nil {
		utils.SendBadRequest(w, err, "error reading login request json")
		return
	}

	user, err := s.gophermart.Login(r.Context(), loginReq)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			utils.SendErrorMsg(w, err, "", http.StatusUnauthorized)
		}
		domain.SendError(w, err)
	}

	cookie, err := s.auth.CreateTokenCookie(user)
	if err != nil {
		utils.SendInternalError(w, err, "error creating token")
	}
	http.SetCookie(w, cookie)
}
