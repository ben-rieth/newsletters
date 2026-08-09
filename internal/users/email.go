package users

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var InvalidEmailError = errors.New("Invalid email")

// mail.ParseAddress also accepts display-name forms like `Foo <a@b.com>`, which
// must never reach storage or a mail provider's recipient field.
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// CanonicalizeEmail returns the single form of an address used for storage,
// whitelist lookups and uniqueness. Every comparison in the system is exact, so
// they all have to agree on that form.
func CanonicalizeEmail(raw string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil {
		return "", InvalidEmailError
	}

	canonical := strings.ToLower(addr.Address)
	if !emailPattern.MatchString(canonical) {
		return "", InvalidEmailError
	}

	return canonical, nil
}
