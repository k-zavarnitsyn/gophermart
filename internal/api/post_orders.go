package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

func (s *gophermartServer) PostOrder(w http.ResponseWriter, r *http.Request) {
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		utils.SendInternalError(w, err, "error reading request body")
	}
	defer utils.CloseWithLogging(r.Body)

	if len(reqData) == 0 {
		utils.SendBadRequest(w, domain.ErrBadRequest, "request data is empty")
	}
	id, err := uuid.NewV6()
	if err != nil {
		utils.SendInternalError(w, err, "error generating uuid")
	}
	authData := auth.FromContext(r.Context())
	err = s.gophermart.PostOrder(r.Context(), &entity.Order{
		ID:     id,
		UserID: authData.UserID,
		Number: string(reqData),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBadOrderNumber):
			utils.SendDomainError(w, err, http.StatusUnprocessableEntity)
		case errors.Is(err, domain.ErrOrderCreatedByCurrentUser):
			utils.SendDomainError(w, err, http.StatusOK)
		case errors.Is(err, domain.ErrOrderCreatedByOtherUser):
			utils.SendDomainError(w, err, http.StatusConflict)
		default:
			domain.SendError(w, err)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
