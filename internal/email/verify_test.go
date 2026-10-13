package email

import (
	"testing"

	"github.com/ben-rieth/newsletter-api/internal/auth"
	"github.com/ben-rieth/newsletter-api/internal/config"
)

func newBudgetTestService(t *testing.T) *EmailVerifyService {
	return NewEmailVerifyService(t.Context(), nil, nil, config.Config{}, nil)
}

func trySend(s *EmailVerifyService, recipient string) bool {
	if !s.hasSendBudget(recipient) {
		return false
	}

	s.spendSendBudget(recipient)
	return true
}

func TestSendBudgetEnforcesCooldown(t *testing.T) {
	s := newBudgetTestService(t)

	if !trySend(s, "user@example.com") {
		t.Fatal("first send was refused")
	}

	if trySend(s, "user@example.com") {
		t.Error("immediate resend was allowed")
	}
}

func TestSendBudgetIsPerRecipient(t *testing.T) {
	s := newBudgetTestService(t)

	trySend(s, "user@example.com")

	if !trySend(s, "other@example.com") {
		t.Error("one recipient's cooldown blocked another")
	}
}

// A send that fails after the check never reaches the spend, so it must leave
// the budget untouched for the retry.
func TestCheckingSendBudgetSpendsNothing(t *testing.T) {
	s := newBudgetTestService(t)

	const recipient = "user@example.com"

	for range 3 {
		if !s.hasSendBudget(recipient) {
			t.Fatal("checking the budget spent from it")
		}
	}

	if !trySend(s, recipient) {
		t.Error("send after failed attempts was refused")
	}
}

// The cooldown limiter is swapped for one that never refills, and resetting it
// stands in for the cooldown elapsing. Refused attempts are made between every
// send: if they spent from the daily budget, it would run out before the cap.
func TestSendBudgetCapsDailySendsAndRefusalsCostNothing(t *testing.T) {
	s := newBudgetTestService(t)
	cooldown := auth.NewKeyedLimiter(t.Context(), 0, 1)
	s.sendCooldown = cooldown

	const recipient = "user@example.com"

	for i := range maxVerificationSendsPerWindow {
		cooldown.Reset(recipient)

		if !trySend(s, recipient) {
			t.Fatalf("send %d of %d was refused", i+1, maxVerificationSendsPerWindow)
		}

		for range 3 {
			if trySend(s, recipient) {
				t.Fatalf("send %d: resend inside the cooldown was allowed", i+1)
			}
		}
	}

	cooldown.Reset(recipient)

	if trySend(s, recipient) {
		t.Error("send past the daily cap was allowed")
	}
}
