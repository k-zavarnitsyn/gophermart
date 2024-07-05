package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	ContentType         = "Content-Type"
	ContentTypeText     = "text/plain"
	ContentTypeTextHTML = "text/html"
	ContentTypeJSON     = "application/json"
	ContentTypeJSONUTF8 = "application/json; charset=utf-8"
	ContentEncoding     = "Content-Encoding"
	AcceptEncoding      = "Accept-Encoding"
)

func JSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set(ContentType, ContentTypeJSONUTF8)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, err := fmt.Fprintf(w, `{"error": "%s"}`, msg)
	if err != nil {
		log.WithError(err).WithField("msg", msg).Error("unable to write json error")
	}
}

func SendDomainError(w http.ResponseWriter, err error, status int) {
	SendErrorMsg(w, err, err.Error(), status)
}

func SendErrorMsg(w http.ResponseWriter, err error, msg string, status int) {
	log.WithError(err).Info(msg)
	JSONError(w, msg, status)
}

func SendBadRequest(w http.ResponseWriter, err error, msg string) {
	SendErrorMsg(w, err, msg, http.StatusBadRequest)
}

func SentNotFound(w http.ResponseWriter, err error, msg string) {
	SendErrorMsg(w, err, msg, http.StatusNotFound)
}

func SendInternalError(w http.ResponseWriter, err error, msg string) {
	log.WithError(err).Error(msg)
	JSONError(w, msg, http.StatusInternalServerError)
}

func SendResponse(w http.ResponseWriter, obj any, status int) {
	w.Header().Set(ContentType, ContentTypeJSONUTF8)
	w.WriteHeader(status)
	data, err := json.Marshal(obj)
	if err != nil {
		SendInternalError(w, err, "unable to marshal object")
		return
	}
	if _, err := w.Write(data); err != nil {
		SendInternalError(w, err, "unable to write response")
		return
	}
}
