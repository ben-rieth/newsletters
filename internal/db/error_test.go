package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// Callers turn a true here into a 409 and anything else into a 500, so a wrong
// code would silently report every duplicate signup as a server fault.
func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "unique violation", err: &pgconn.PgError{Code: "23505"}, want: true},
		{
			name: "wrapped unique violation",
			err:  fmt.Errorf("saving user: %w", &pgconn.PgError{Code: "23505"}),
			want: true,
		},
		{name: "foreign key violation", err: &pgconn.PgError{Code: "23503"}, want: false},
		{name: "not null violation", err: &pgconn.PgError{Code: "23502"}, want: false},
		{name: "check violation", err: &pgconn.PgError{Code: "23514"}, want: false},
		{name: "non postgres error", err: errors.New("dial tcp: connection refused"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUniqueViolation(tt.err); got != tt.want {
				t.Errorf("IsUniqueViolation(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
