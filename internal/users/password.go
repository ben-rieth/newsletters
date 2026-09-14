package users

import (
	"errors"
	"unicode/utf8"
)

const MinPasswordLength = 12

// bcrypt silently truncates past this, so a longer password's tail would
// contribute nothing to the hash.
const MaxPasswordBytes = 72

var PasswordTooShortError = errors.New("Password must be at least 12 characters")
var PasswordTooLongError = errors.New("Password must be at most 72 bytes")

func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return PasswordTooShortError
	}

	if len(password) > MaxPasswordBytes {
		return PasswordTooLongError
	}

	return nil
}
