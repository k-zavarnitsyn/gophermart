package api

import (
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

func (s *gophermartServer) GetOrders(w http.ResponseWriter, r *http.Request) {
	clientData := auth.FromContext(r.Context())
	orders, err := s.gophermart.GetOrders(r.Context(), clientData.UserID)
	if err != nil {
		domain.SendError(w, err)
		return
	}

	if len(orders) == 0 {
		utils.SendResponse(w, []struct{}{}, http.StatusNoContent)
	} else {
		ordersResp := make([]entity.OrderResponse, len(orders))
		for i, o := range orders {
			ordersResp[i] = entity.OrderResponse{
				Number:    o.Number,
				Status:    o.Status,
				CreatedAt: o.CreatedAt,
				Accrual:   o.Accrual,
			}
		}
		utils.SendResponse(w, ordersResp, http.StatusOK)
	}
}
