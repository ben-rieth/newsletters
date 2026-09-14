package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSessionIsCurrent(t *testing.T) {
	cutoff := time.Date(2026, 9, 10, 12, 0, 30, 500_000_000, time.UTC)

	tests := []struct {
		name     string
		issuedAt time.Time
		want     bool
	}{
		{
			name:     "issued well after the cutoff",
			issuedAt: cutoff.Add(time.Minute),
			want:     true,
		},
		{
			name:     "issued well before the cutoff",
			issuedAt: cutoff.Add(-time.Minute),
			want:     false,
		},
		{
			// Where the token a password change hands back lands: stamped with a
			// truncated second that falls below the sub-second cutoff it followed.
			name:     "issued in the cutoff's own second, before it",
			issuedAt: cutoff.Truncate(time.Second),
			want:     true,
		},
		{
			name:     "issued in the cutoff's own second, after it",
			issuedAt: cutoff.Add(100 * time.Millisecond),
			want:     true,
		},
		{
			name:     "issued in the previous second",
			issuedAt: cutoff.Truncate(time.Second).Add(-time.Second),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SessionIsCurrent(tt.issuedAt, cutoff); got != tt.want {
				t.Errorf("SessionIsCurrent = %v, want %v", got, tt.want)
			}
		})
	}
}

// The same-second allowance in SessionIsCurrent exists only because of this
// truncation, so pin it: were the library to stop, that could be tightened.
func TestIssuedAtIsTruncatedToSeconds(t *testing.T) {
	token, err := GenerateToken("user-1", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}

	if got := claims.IssuedAt.Time.Nanosecond(); got != 0 {
		t.Errorf("IssuedAt carries sub-second precision (%d ns); the cutoff comparison assumes it does not", got)
	}
}

func TestFreshTokenSurvivesCutoffSetJustBefore(t *testing.T) {
	cutoff := time.Now()

	token, err := GenerateToken("user-1", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}

	if !SessionIsCurrent(claims.IssuedAt.Time, cutoff) {
		t.Error("a token issued after the cutoff was treated as revoked")
	}
}

func TestTokenIssuedBeforeCutoffIsRejected(t *testing.T) {
	issuedAt := jwt.NewNumericDate(time.Now().Add(-time.Hour)).Time

	if SessionIsCurrent(issuedAt, time.Now()) {
		t.Error("a token issued an hour before the cutoff was treated as current")
	}
}
