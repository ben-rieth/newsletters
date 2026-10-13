package handler

import (
	"context"
	"reflect"
	"testing"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"golang.org/x/crypto/bcrypt"
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

const testPassword = "correct horse battery"

func newPasswordTestHandler(t *testing.T) *AuthHandler {
	return &AuthHandler{signInLimiter: NewSignInLimiter(t.Context())}
}

func newPasswordTestUser(t *testing.T, id string) db.AppUser {
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	return db.AppUser{ID: id, Password: string(hash)}
}

func fromAddress(t *testing.T, clientKey string) context.Context {
	return context.WithValue(t.Context(), auth.ClientKeyKey, clientKey)
}

func mustCheckPassword(t *testing.T, h *AuthHandler, ctx context.Context, user db.AppUser, password string) bool {
	t.Helper()

	ok, err := checkPassword(ctx, h.signInLimiter, user, password)
	if err != nil {
		t.Fatalf("checkPassword: %v", err)
	}

	return ok
}

func spendGuesses(t *testing.T, h *AuthHandler, ctx context.Context, user db.AppUser, n int) {
	t.Helper()

	for range n {
		mustCheckPassword(t, h, ctx, user, "wrong password")
	}
}

func TestCheckPasswordAcceptsCorrectPassword(t *testing.T) {
	h := newPasswordTestHandler(t)
	user := newPasswordTestUser(t, "user-1")

	if !mustCheckPassword(t, h, fromAddress(t, "203.0.113.9"), user, testPassword) {
		t.Error("correct password was refused")
	}

	if mustCheckPassword(t, h, fromAddress(t, "203.0.113.9"), user, "wrong password") {
		t.Error("wrong password was accepted")
	}
}

// Once the budget is gone the password is not even compared, or a guesser
// could keep confirming hits after the limit.
func TestCheckPasswordRefusesCorrectPasswordOnceBudgetSpent(t *testing.T) {
	h := newPasswordTestHandler(t)
	user := newPasswordTestUser(t, "user-1")
	ctx := fromAddress(t, "203.0.113.9")

	spendGuesses(t, h, ctx, user, signInGuessBurst)

	if mustCheckPassword(t, h, ctx, user, testPassword) {
		t.Error("correct password was accepted past the guess budget")
	}
}

// The whole point of keying by account and address: a stranger burning their
// budget must not lock the owner out.
func TestCheckPasswordBudgetDoesNotSpillToOtherAddresses(t *testing.T) {
	h := newPasswordTestHandler(t)
	user := newPasswordTestUser(t, "user-1")

	spendGuesses(t, h, fromAddress(t, "198.51.100.7"), user, signInGuessBurst)

	if !mustCheckPassword(t, h, fromAddress(t, "203.0.113.9"), user, testPassword) {
		t.Error("owner on another address was locked out by someone else's guesses")
	}
}

func TestCheckPasswordBudgetDoesNotSpillToOtherAccounts(t *testing.T) {
	h := newPasswordTestHandler(t)
	ctx := fromAddress(t, "203.0.113.9")

	spendGuesses(t, h, ctx, newPasswordTestUser(t, "user-1"), signInGuessBurst)

	if !mustCheckPassword(t, h, ctx, newPasswordTestUser(t, "user-2"), testPassword) {
		t.Error("another account on the same address was locked out")
	}
}

func TestCheckPasswordSuccessRestoresBudget(t *testing.T) {
	h := newPasswordTestHandler(t)
	user := newPasswordTestUser(t, "user-1")
	ctx := fromAddress(t, "203.0.113.9")

	// Two rounds of burst-1 wrong guesses add up to more than one burst, so the
	// second success only happens if the first one reset the budget.
	for round := range 2 {
		spendGuesses(t, h, ctx, user, signInGuessBurst-1)

		if !mustCheckPassword(t, h, ctx, user, testPassword) {
			t.Fatalf("round %d: correct password was refused", round+1)
		}
	}
}

func TestCheckPasswordFailsClosedWithoutClientKey(t *testing.T) {
	h := newPasswordTestHandler(t)
	user := newPasswordTestUser(t, "user-1")

	ok, err := checkPassword(t.Context(), h.signInLimiter, user, testPassword)
	if err == nil {
		t.Error("expected an error when the client key is missing")
	}

	if ok {
		t.Error("password was accepted without a client key")
	}
}
