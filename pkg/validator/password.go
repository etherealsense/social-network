package validator

import "errors"

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong  = errors.New("password must be less than 64 characters long")
	ErrPasswordEmpty    = errors.New("password cannot be empty")
)

func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordEmpty
	}

	if len(password) > 64 {
		return ErrPasswordTooLong
	}

	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}
