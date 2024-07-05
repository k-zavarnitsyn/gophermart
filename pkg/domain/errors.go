package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/k-zavarnitsyn/gophermart/internal/utils"
)

var ErrNotFound = NewError("not found")
var ErrBadRequest = NewError("bad request")
var ErrAuthentication = NewError("authentication error")
var ErrBadOrderNumber = fmt.Errorf("%w: bad order number format", ErrBadRequest)
var ErrOrderNumberTooLong = fmt.Errorf("%w: order number too long", ErrBadOrderNumber)
var ErrOrderNumberExists = NewError("order number already exists")
var ErrOrderCreatedByCurrentUser = fmt.Errorf("%w: created by current user", ErrOrderNumberExists)
var ErrOrderCreatedByOtherUser = fmt.Errorf("%w: created by other user", ErrOrderNumberExists)
var ErrNotEnoughAccruals = NewError("insufficient funds in the account")
var ErrRegister = NewError("register error")
var ErrLoginExists = fmt.Errorf("%w: login already exists", ErrRegister)
var ErrInvalidToken = fmt.Errorf("%w: invalid token", ErrAuthentication)
var ErrTokenNotProvided = fmt.Errorf("%w: no token provided", ErrAuthentication)
var ErrUserIDNotProvided = fmt.Errorf("%w: no token user ID provided", ErrBadRequest)
var ErrUserLoginNotProvided = fmt.Errorf("%w: no token user login provided", ErrBadRequest)
var ErrTokenExpirationNotProvided = fmt.Errorf("%w: no token expiration time provided", ErrBadRequest)

type Err struct {
	err error
}

func NewError(format string, args ...any) *Err {
	return &Err{err: fmt.Errorf(format, args...)}
}

func (e *Err) Error() string {
	return e.err.Error()
}

func (e *Err) Unwrap() error {
	return e.err
}

func SendError(w http.ResponseWriter, err error, msg ...string) {
	if msg != nil {
		err = fmt.Errorf(strings.Join(msg, ": ")+": %w", err)
	}
	var e *Err
	if errors.As(err, &e) {
		switch {
		case errors.Is(err, ErrNotFound):
			utils.SendDomainError(w, err, http.StatusNotFound)
		case errors.Is(err, ErrAuthentication):
			utils.SendDomainError(w, err, http.StatusUnauthorized)
		default:
			utils.SendDomainError(w, err, http.StatusBadRequest)
		}
	} else {
		utils.SendInternalError(w, err, err.Error())
	}
}
