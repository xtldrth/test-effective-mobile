package internal

import "errors"

var ErrInternal = errors.New("internal error")

type ErrNotFound struct {
	Resource string
}

func (e ErrNotFound) Error() string {
	return e.Resource + " not found"
}

type ErrValidation struct {
	Msg string
}

func (e ErrValidation) Error() string {
	return e.Msg
}
