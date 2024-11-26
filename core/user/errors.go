package user

import (
	"errors"
	"fmt"
)

var ErrNoUserInformation = errors.New("no user information")

type NotFoundError struct {
	Email string
}

func (e NotFoundError) Error() string {
	cause := "could not find user"
	if e.Email != "" {
		cause += fmt.Sprintf(" with email \"%s\"", e.Email)
	}
	return cause
}

type DuplicateRecordError struct {
	Email string
}

func (e DuplicateRecordError) Error() string {
	cause := "duplicate user"
	if e.Email != "" {
		cause += fmt.Sprintf(" with email \"%s\"", e.Email)
	}
	return cause
}

type InvalidError struct {
	Email string
}

func (e InvalidError) Error() string {
	return fmt.Sprintf("empty field with email %q", e.Email)
}
