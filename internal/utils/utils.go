package utils

import (
	"encoding/json"
	"io"

	log "github.com/sirupsen/logrus"
)

type Factory[T any] interface {
	Create() T
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func ContainsWhere[T comparable](s []T, predicate func(e T) bool) bool {
	for _, v := range s {
		if predicate(v) {
			return true
		}
	}
	return false
}

func FromPointer[T any](value *T) T {
	var result T
	if value == nil {
		return result
	}
	result = *value

	return result
}

func ToPointer[T any](value T) *T {
	return &value
}

func CloseWithLogging(closeFunc io.Closer, errorMsg ...string) {
	if err := closeFunc.Close(); err != nil {
		log.WithError(err).Error(errorMsg)
	}
}

func ReadJSON[T any](r io.Reader) (*T, error) {
	var obj *T
	err := json.NewDecoder(r).Decode(obj)

	return obj, err
}
