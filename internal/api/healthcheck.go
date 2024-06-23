package api

import (
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/utils"
)

func (s *gophermartServer) Healthcheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.dbPinger.Ping(ctx); err != nil {
		utils.SendInternalError(w, err, "server healthcheck failed")
		return
	}

	w.WriteHeader(http.StatusOK)
}
