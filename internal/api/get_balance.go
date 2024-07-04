package api

import (
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
)

func (s *gophermartServer) GetBalance(w http.ResponseWriter, r *http.Request) {
	clientData := auth.FromContext(r.Context())
	balance, err := s.gophermart.GetBalance(r.Context(), clientData.UserID)
	if err != nil {
		domain.SendError(w, err)
		return
	}

	utils.SendResponse(w, balance, http.StatusOK)
}
