package validator

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmailInvalid = errors.New("invalid email format")
	ErrEmailEmpty   = errors.New("email cannot be empty")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return ErrEmailEmpty
	}

	if len(email) > 64 {
		return ErrEmailInvalid
	}

	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}

	return nil
}
