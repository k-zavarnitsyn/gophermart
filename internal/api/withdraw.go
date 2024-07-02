package api

import (
	"errors"
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

func (s *gophermartServer) Withdraw(w http.ResponseWriter, r *http.Request) {
	clientData := auth.FromContext(r.Context())
	reqData, err := utils.ReadJSON[entity.WithdrawRequest](r.Body)
	if err != nil {
		utils.SendBadRequest(w, err, "error reading withdraw request json")
		return
	}
	err = s.gophermart.Withdraw(r.Context(), &entity.Withdraw{
		UserID:      clientData.UserID,
		OrderNumber: reqData.OrderNumber,
		Value:       reqData.Sum,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotEnoughAccruals) {
			utils.SendDomainError(w, err, http.StatusPaymentRequired)
		} else if errors.Is(err, domain.ErrBadOrderNumber) {
			utils.SendDomainError(w, err, http.StatusUnprocessableEntity)
		} else {
			utils.SendError(w, err)
		}
		return
	}
}
