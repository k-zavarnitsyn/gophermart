package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	log "github.com/sirupsen/logrus"
)

func JSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, err := fmt.Fprintf(w, `{"error": "%s"}`, msg)
	if err != nil {
		log.WithError(err).WithField("msg", msg).Error("unable to write json error")
	}
}

func SendError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.DomainError) {
		if errors.Is(err, domain.ErrNotFound) {
			SendDomainError(w, err, http.StatusNotFound)
		} else {
			SendDomainError(w, err, http.StatusBadRequest)
		}
	} else {
		SendInternalError(w, err, err.Error())
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

func SendResponse(w http.ResponseWriter, obj any) {
	data, err := json.Marshal(obj)
	if err != nil {
		SendInternalError(w, err, "unable to marshal object")
	}
	if _, err := w.Write(data); err != nil {
		SendInternalError(w, err, "unable to write response")
	}
}
