package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/k-zavarnitsyn/gophermart/internal/api"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
)

func WithGzipResponse(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(api.AcceptEncoding), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			utils.SendInternalError(w, err, "unable to initialize gzip writer")
			return
		}
		defer utils.CloseWithLogging(gz, "unable to close gzip writer")

		customWriter := &CustomWriter{ResponseWriter: w, Writer: gz}
		customWriter.Header().Set(api.ContentEncoding, "gzip")
		h.ServeHTTP(customWriter, r)
	})
}

func WithGzipRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(api.ContentEncoding) != "gzip" {
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			utils.SendInternalError(w, err, "unable to initialize gzip reader")
			return
		}
		defer utils.CloseWithLogging(gz, "unable to close gzip reader")
		r.Body = gz

		h.ServeHTTP(w, r)
	})
}
