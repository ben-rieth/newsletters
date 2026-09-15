package users

import (
	"errors"
	"testing"
)

func TestCanonicalizeEmailAccepts(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain address", raw: "user@example.com", want: "user@example.com"},
		{name: "uppercase is lowered", raw: "User@Example.COM", want: "user@example.com"},
		{name: "surrounding whitespace is trimmed", raw: "  user@example.com\t", want: "user@example.com"},
		// Subaddresses stay distinct, because every comparison in the system is
		// exact and the provider decides what routes where.
		{name: "plus addressing is preserved", raw: "user+news@example.com", want: "user+news@example.com"},
		{name: "dots are preserved", raw: "first.last@example.com", want: "first.last@example.com"},
		{name: "subdomain", raw: "user@mail.example.co.uk", want: "user@mail.example.co.uk"},
		{name: "two letter tld", raw: "user@example.io", want: "user@example.io"},
		{name: "hyphenated domain", raw: "user@my-example.com", want: "user@my-example.com"},

		// Display names and comments are parsed off rather than rejected, so what
		// reaches storage and a provider's recipient field is the bare address.
		{name: "display name form is reduced", raw: "Foo <user@example.com>", want: "user@example.com"},
		{name: "angle brackets are stripped", raw: "<User@Example.com>", want: "user@example.com"},
		{name: "trailing comment is dropped", raw: "user@example.com (Foo)", want: "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeEmail(tt.raw)
			if err != nil {
				t.Fatalf("CanonicalizeEmail(%q): %v", tt.raw, err)
			}

			if got != tt.want {
				t.Errorf("CanonicalizeEmail(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCanonicalizeEmailRejects(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "no at sign", raw: "not-an-email"},
		{name: "no domain", raw: "user@"},
		{name: "no local part", raw: "@example.com"},
		{name: "bare hostname with no tld", raw: "user@example"},
		{name: "single character tld", raw: "user@example.c"},
		{name: "numeric tld", raw: "user@example.12"},
		// mail.ParseAddress accepts these, so the pattern is the only thing keeping
		// them out of storage and out of a provider's recipient field.
		{name: "quoted local part with a space", raw: `"first last"@example.com`},
		{name: "two addresses", raw: "user@example.com, other@example.com"},
		{name: "newline injection", raw: "user@example.com\nBcc: other@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeEmail(tt.raw)

			if !errors.Is(err, InvalidEmailError) {
				t.Errorf("CanonicalizeEmail(%q) = %q, %v; want InvalidEmailError", tt.raw, got, err)
			}

			if got != "" {
				t.Errorf("CanonicalizeEmail(%q) returned %q alongside an error", tt.raw, got)
			}
		})
	}
}

// Storage, whitelist lookups and uniqueness all compare the canonical form, so
// canonicalizing an already-canonical address has to be a no-op.
func TestCanonicalizeEmailIsIdempotent(t *testing.T) {
	for _, raw := range []string{"User@Example.COM", "  user+news@example.com  ", "First.Last@Mail.Example.Co.Uk"} {
		once, err := CanonicalizeEmail(raw)
		if err != nil {
			t.Fatalf("CanonicalizeEmail(%q): %v", raw, err)
		}

		twice, err := CanonicalizeEmail(once)
		if err != nil {
			t.Fatalf("CanonicalizeEmail(%q): %v", once, err)
		}

		if once != twice {
			t.Errorf("canonicalizing %q twice gave %q then %q", raw, once, twice)
		}
	}
}
