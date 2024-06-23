package domain

import "errors"

var DomainError = NewError("domain error")
var ErrNotFound = NewError("not found")
var ErrBadRequest = NewError("bad request")
var ErrLoginExists = NewError("login already exists")

type Error struct {
	Err error
}

func NewError(text string) error {
	return &Error{
		Err: errors.New(text),
	}
}

func (d Error) Error() string {
	return d.Err.Error()
}

func (d Error) Is(err error) bool {
	// быстрый способ понять, что это доменная ошибка
	if err == DomainError {
		return true
	}

	return errors.Is(err, d.Err)
}

// func (d Error) Unwrap() error {
// 	return d.Err
// }
