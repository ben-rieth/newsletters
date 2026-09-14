package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "a-test-secret-that-is-long-enough-to-pass"

func TestParseTokenAcceptsOwnToken(t *testing.T) {
	token, err := GenerateToken("user-1", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken rejected a token we just issued: %v", err)
	}

	if claims.Subject != "user-1" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "user-1")
	}

	if claims.IssuedAt == nil {
		t.Error("IssuedAt is nil, so the session cutoff check could not run")
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	token, err := GenerateToken("user-1", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if _, err := ParseToken(token, "a-different-secret-of-adequate-length"); err == nil {
		t.Error("ParseToken accepted a token signed with a different secret")
	}
}

func signed(t *testing.T, method jwt.SigningMethod, key any, claims jwt.Claims) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}

	return token
}

func validClaims() jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Subject:   "user-1",
		Issuer:    tokenIssuer,
		Audience:  jwt.ClaimStrings{tokenAudience},
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
}

func TestParseTokenRejectsAlgNone(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{RegisteredClaims: validClaims()}).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("signing unsigned token: %v", err)
	}

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted an unsigned token")
	}
}

// HS512 verifies against the same secret, so only pinning the algorithm keeps a
// token the issuer would never mint from being honoured.
func TestParseTokenRejectsOtherHMACFamilies(t *testing.T) {
	token := signed(t, jwt.SigningMethodHS512, []byte(testSecret), Claims{RegisteredClaims: validClaims()})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted an HS512 token")
	}
}

func TestParseTokenRejectsMissingExpiry(t *testing.T) {
	claims := validClaims()
	claims.ExpiresAt = nil

	token := signed(t, jwt.SigningMethodHS256, []byte(testSecret), Claims{RegisteredClaims: claims})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted a token that never expires")
	}
}

func TestParseTokenRejectsExpired(t *testing.T) {
	claims := validClaims()
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))

	token := signed(t, jwt.SigningMethodHS256, []byte(testSecret), Claims{RegisteredClaims: claims})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted an expired token")
	}
}

func TestParseTokenRejectsWrongIssuer(t *testing.T) {
	claims := validClaims()
	claims.Issuer = "somebody-else"

	token := signed(t, jwt.SigningMethodHS256, []byte(testSecret), Claims{RegisteredClaims: claims})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted a token from another issuer")
	}
}

func TestParseTokenRejectsWrongAudience(t *testing.T) {
	claims := validClaims()
	claims.Audience = jwt.ClaimStrings{"some-other-api"}

	token := signed(t, jwt.SigningMethodHS256, []byte(testSecret), Claims{RegisteredClaims: claims})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted a token meant for another audience")
	}
}

func TestParseTokenRejectsEmptySubject(t *testing.T) {
	claims := validClaims()
	claims.Subject = ""

	token := signed(t, jwt.SigningMethodHS256, []byte(testSecret), Claims{RegisteredClaims: claims})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted a token with no subject")
	}
}

// Without IssuedAt there is nothing to compare against the session cutoff, so
// the token would be unrevokable.
func TestParseTokenRejectsMissingIssuedAt(t *testing.T) {
	claims := validClaims()
	claims.IssuedAt = nil

	token := signed(t, jwt.SigningMethodHS256, []byte(testSecret), Claims{RegisteredClaims: claims})

	if _, err := ParseToken(token, testSecret); err == nil {
		t.Error("ParseToken accepted a token with no issued-at")
	}
}
