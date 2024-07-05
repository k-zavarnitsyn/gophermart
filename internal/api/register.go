package api

import (
	"errors"
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

func (s *gophermartServer) Register(w http.ResponseWriter, r *http.Request) {
	reqData, err := utils.ReadJSON[entity.RegisterRequest](r.Body)
	if err != nil {
		utils.SendBadRequest(w, err, "error reading register request json")
		return
	}
	user, err := s.gophermart.Register(r.Context(), reqData)
	if err != nil {
		if errors.Is(err, domain.ErrLoginExists) {
			utils.SendDomainError(w, err, http.StatusConflict)
		} else {
			domain.SendError(w, err, "error registering user")
		}
		return
	}

	cookie, err := s.auth.CreateTokenCookie(user)
	if err != nil {
		domain.SendError(w, err, "error creating token")
		return
	}
	http.SetCookie(w, cookie)
}
