package api

import (
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

func (s *gophermartServer) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	clientData := auth.FromContext(r.Context())
	withdrawals, err := s.gophermart.GetWithdrawals(r.Context(), clientData.UserID)
	if err != nil {
		utils.SendError(w, err)
		return
	}

	if len(withdrawals) == 0 {
		utils.SendResponse(w, []struct{}{}, http.StatusNoContent)
	} else {
		withdrawalsResp := make([]entity.WithdrawalsResponse, len(withdrawals))
		for i, withdraw := range withdrawals {
			withdrawalsResp[i] = entity.WithdrawalsResponse{
				OrderNumber: withdraw.OrderNumber,
				Sum:         withdraw.Value,
				ProcessedAt: withdraw.CreatedAt,
			}
		}
		utils.SendResponse(w, withdrawalsResp, http.StatusOK)
	}
}
