package users

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "empty", password: "", wantErr: PasswordTooShortError},
		{name: "one under the minimum", password: strings.Repeat("a", MinPasswordLength-1), wantErr: PasswordTooShortError},
		{name: "exactly the minimum", password: strings.Repeat("a", MinPasswordLength), wantErr: nil},
		{name: "exactly the byte ceiling", password: strings.Repeat("a", MaxPasswordBytes), wantErr: nil},
		{name: "one byte over the ceiling", password: strings.Repeat("a", MaxPasswordBytes+1), wantErr: PasswordTooLongError},

		// The minimum counts runes, so a short multi-byte password is judged by
		// what the user typed rather than by how it encodes.
		{name: "minimum multi-byte runes", password: strings.Repeat("パ", MinPasswordLength), wantErr: nil},
		{name: "one multi-byte rune short", password: strings.Repeat("パ", MinPasswordLength-1), wantErr: PasswordTooShortError},

		// The ceiling counts bytes, because that is what bcrypt truncates.
		{name: "eighteen four-byte runes fit", password: strings.Repeat("🔒", 18), wantErr: nil},
		{name: "nineteen four-byte runes do not", password: strings.Repeat("🔒", 19), wantErr: PasswordTooLongError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidatePassword(%d runes, %d bytes) = %v, want %v",
					len([]rune(tt.password)), len(tt.password), err, tt.wantErr)
			}
		})
	}
}
