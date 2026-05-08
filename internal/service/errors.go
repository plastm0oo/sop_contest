package service

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountBlocked        = errors.New("account blocked")
	ErrFeedbackAlreadyExists = errors.New("feedback already exists")
	ErrTeacherNotFound       = errors.New("teacher not found")
)

type ValidationError struct {
	Details map[string]string
}

func (e ValidationError) Error() string {
	return "validation failed"
}
