package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
}

type contextKey string

const ClaimsKey contextKey = "claims"

const AccessTokenTTL time.Duration = time.Hour * 1
const RefreshTokenTTL time.Duration = time.Hour * 24 * 30

const tokenIssuer = "newsletters"
const tokenAudience = "newsletters-api"

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	return claims, ok
}

func GenerateToken(
	userID string,
	tokenSecret string,
) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
			Issuer:    tokenIssuer,
			Audience:  jwt.ClaimStrings{tokenAudience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func ParseToken(tokenString, tokenSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) { return []byte(tokenSecret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithAudience(tokenAudience),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("Token has no subject")
	}

	// Revocation works by comparing IssuedAt against the account's session
	// cutoff, so a token without one could never be revoked.
	if claims.IssuedAt == nil {
		return nil, fmt.Errorf("Token has no issued-at")
	}

	return claims, nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func MakeVerificationToken() (string, error) {
	max := big.NewInt(100000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%08d", n.Int64()), nil
}
