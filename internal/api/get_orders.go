package api

import (
	"net/http"
)

func (s *gophermartServer) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(ContentType, ContentTypeJSON)
}
