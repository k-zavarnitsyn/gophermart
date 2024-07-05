package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/k-zavarnitsyn/gophermart/internal/config"
	log "github.com/sirupsen/logrus"
)

type Logger struct {
	cfg              *config.Log
	WithRequestData  bool
	WithResponseData bool
}

func NewLogger(cfg *config.Log) *Logger {
	return &Logger{
		cfg: cfg,
	}
}

func (l *Logger) WithRequestLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var body []byte
		if l.cfg.WithRequestData && r.Method != http.MethodGet {
			var err error
			body, err = io.ReadAll(r.Body)
			if err != nil {
				log.WithError(err).Error("error reading body")
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		h.ServeHTTP(w, r)
		duration := time.Since(start)

		log.Infoln(r.Method, r.RequestURI, string(body), duration)
	})
}

func (l *Logger) WithResponseLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := newLoggingResponseWriter(w, l.cfg.WithResponseData)
		h.ServeHTTP(lw, r)

		if l.cfg.WithResponseData {
			log.WithField("status", lw.responseData.status).
				WithField("size", lw.responseData.size).
				WithField("data", lw.responseData.data.String()).
				Info()
		} else {
			log.WithField("status", lw.responseData.status).
				WithField("size", lw.responseData.size).
				Info()
		}
	})
}
