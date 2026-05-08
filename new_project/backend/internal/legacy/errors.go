package legacy

import "errors"

var ErrInvalid = errors.New("invalid")

type domainError struct {
	message string
	cause   error
}

func (e domainError) Error() string {
	return e.message
}

func (e domainError) Unwrap() error {
	return e.cause
}

func newInvalidError(message string) error {
	return domainError{
		message: message,
		cause:   ErrInvalid,
	}
}

func newForbiddenError(message string) error {
	return domainError{
		message: message,
		cause:   ErrForbidden,
	}
}
