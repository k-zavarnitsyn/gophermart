package domain

import (
	"errors"
	"fmt"
)

var Error = errors.New("error")
var ErrNotFound = fmt.Errorf("%w: not found", Error)
var ErrBadRequest = fmt.Errorf("%w: bad request", Error)
var ErrBadOrderNumber = fmt.Errorf("%w: bad order number format", ErrBadRequest)
var ErrOrderNumberTooLong = fmt.Errorf("%w: order number too long", ErrBadOrderNumber)
var ErrOrderNumberExists = fmt.Errorf("%w: order number already exists", Error)
var ErrOrderCreatedByCurrentUser = fmt.Errorf("%w: created by current user", ErrOrderNumberExists)
var ErrOrderCreatedByOtherUser = fmt.Errorf("%w: created by other user", ErrOrderNumberExists)
var ErrNotEnoughAccruals = fmt.Errorf("%w: insufficient funds in the account", Error)
var ErrLoginExists = fmt.Errorf("%w: login already exists", Error)
