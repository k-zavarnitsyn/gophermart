package utils

import (
	"time"
)

const AttemptsCount = 3
const DefaultTimeout = time.Second * 1
const TimeoutIncrement = time.Second * 2

type Retriable interface {
	error
	IsRetriable() bool
}

type RetriableError struct {
	Err error
}

func (e *RetriableError) Error() string {
	return e.Err.Error()
}

func (e *RetriableError) IsRetriable() bool {
	// Метод переопределяется, поэтому тут просто какая-то не значимая логика
	t, ok := e.Err.(interface {
		Temporary() bool
	})

	return ok && t.Temporary()
}

func RetryEx(operation func() error, isRetriable func(err error) bool, betweenRetries func(timeout time.Duration)) error {
	timeout := DefaultTimeout
	var err error
	for i := 0; i < AttemptsCount; i++ {
		err = operation()
		if !isRetriable(err) {
			return err
		}
		betweenRetries(timeout)
		timeout += TimeoutIncrement
	}

	return err
}

func RetryWaiting(operation func() error, isRetriable func(err error) bool) error {
	return RetryEx(operation, isRetriable, time.Sleep)
}

func Retry(operation func() error) error {
	return RetryEx(operation, func(err error) bool {
		rErr, ok := err.(Retriable)
		return ok && rErr.IsRetriable()
	}, time.Sleep)
}
