package handler

import (
	"reflect"
	"testing"

	"github.com/ben-rieth/newsletter-api/internal/auth"
)

// The tag on refreshInput has to repeat the cookie name as a literal, because a
// struct tag cannot reference a constant. Renaming one and not the other would
// otherwise fail silently.
func TestRefreshInputCookieTagMatchesCookieName(t *testing.T) {
	field, ok := reflect.TypeOf(refreshInput{}).FieldByName("RefreshToken")
	if !ok {
		t.Fatal("refreshInput has no RefreshToken field")
	}

	if got := field.Tag.Get("cookie"); got != auth.RefreshTokenCookie {
		t.Errorf("cookie tag = %q, want %q", got, auth.RefreshTokenCookie)
	}
}
