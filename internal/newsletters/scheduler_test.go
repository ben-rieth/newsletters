package newsletters

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/google/uuid"
)

type sendAttempt struct {
	subject        string
	sender         string
	recipient      string
	body           string
	idempotencyKey string
}

type fakeEmailService struct {
	mu       sync.Mutex
	attempts []sendAttempt

	// errs[i] is returned on attempt i+1. Attempts past the end succeed.
	errs   []error
	onSend func(attempt int)

	assembled        string
	assembleErr      error
	assembleArgs     map[string]any
	assembleTemplate string
}

func (f *fakeEmailService) SendIdempotent(
	ctx context.Context,
	subject, sender, recipient, body, idempotencyKey string,
) (*email.SendResult, error) {
	f.mu.Lock()
	index := len(f.attempts)
	f.attempts = append(f.attempts, sendAttempt{subject, sender, recipient, body, idempotencyKey})
	f.mu.Unlock()

	if f.onSend != nil {
		f.onSend(index + 1)
	}

	if index < len(f.errs) && f.errs[index] != nil {
		return nil, f.errs[index]
	}

	return &email.SendResult{
		ID:   fmt.Sprintf("send-%d", index+1),
		Time: time.Date(2026, time.March, 10, 9, 31, 0, 0, time.UTC),
	}, nil
}

func (f *fakeEmailService) AssembleEmail(templateName string, arguments map[string]any) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.assembleTemplate = templateName
	f.assembleArgs = arguments

	return f.assembled, f.assembleErr
}

func (f *fakeEmailService) Send(context.Context, string, string, string, string) (*email.SendResult, error) {
	panic("Send is not part of the newsletter send path")
}

func (f *fakeEmailService) BackgroundSend(context.Context, string, string, string, string) {
	panic("BackgroundSend is not part of the newsletter send path")
}

func (f *fakeEmailService) sendAttempts() []sendAttempt {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]sendAttempt(nil), f.attempts...)
}

func schedulerWith(emailService email.EmailService, feedService feedFetcher) *Scheduler {
	return &Scheduler{
		emailService: emailService,
		feedService:  feedService,
		cfg: &config.Config{
			WebURL:                "https://web.example.com",
			PublicApiURL:          "https://api.example.com",
			NewsletterSenderEmail: mail.Address{Name: "Slowfeed", Address: "news@example.com"},
		},
		schedulerConfig: &SchedulerConfig{SendRetryDelay: 0},
	}
}

func sendableNewsletter() *SendableNewsletter {
	return &SendableNewsletter{
		ID:               "nl-1",
		Name:             "Weekly Digest",
		Email:            "reader@example.com",
		UserID:           "user-1",
		UnsubscribeToken: "unsub-token",
		LastSendTime:     time.Date(2026, time.March, 3, 9, 30, 0, 0, time.UTC),
		NextSendTime:     time.Date(2026, time.March, 10, 9, 30, 0, 0, time.UTC),
	}
}

func feedViewWithItems(itemIDs ...string) feeds.FeedView {
	items := make([]feeds.FeedItemView, 0, len(itemIDs))
	for _, id := range itemIDs {
		items = append(items, feeds.FeedItemView{ItemID: id})
	}

	return feeds.FeedView{Items: items}
}

func TestGenerateTokensForItems(t *testing.T) {
	views := []feeds.FeedView{
		feedViewWithItems("item-1", "item-2"),
		feedViewWithItems("item-3"),
		feedViewWithItems(),
	}

	tokens := generateTokensForItems(views)

	if len(tokens) != 3 {
		t.Fatalf("got %d tokens, want 3", len(tokens))
	}

	seen := make(map[string]string, len(tokens))
	for itemID, token := range tokens {
		if err := uuid.Validate(token); err != nil {
			// The redirect handler validates the token as a uuid before looking it
			// up, so a non-uuid token here would only ever reach /bad-link.
			t.Errorf("token for %s is not a uuid: %v", itemID, err)
		}

		if other, ok := seen[token]; ok {
			t.Errorf("items %s and %s share token %s", other, itemID, token)
		}
		seen[token] = itemID
	}
}

func TestWrapItemURLs(t *testing.T) {
	views := []feeds.FeedView{
		feedViewWithItems("item-1", "item-2"),
		feedViewWithItems("item-3"),
	}

	tokens := map[string]string{
		"item-1": "token-1",
		"item-2": "token-2",
		"item-3": "token-3",
	}

	got := wrapItemURLs(views, tokens, "https://api.example.com")

	for _, view := range got {
		for _, item := range view.Items {
			want := "https://api.example.com/link/" + tokens[item.ItemID]
			if item.TrackingURL != want {
				t.Errorf("TrackingURL for %s = %q, want %q", item.ItemID, item.TrackingURL, want)
			}
		}
	}
}

// The returned views are the argument, so a caller reading through its own slice
// after the call has to see the wrapped URLs too.
func TestWrapItemURLsRewritesTheGivenViews(t *testing.T) {
	views := []feeds.FeedView{feedViewWithItems("item-1")}

	wrapItemURLs(views, map[string]string{"item-1": "token-1"}, "https://api.example.com")

	if got := views[0].Items[0].TrackingURL; got != "https://api.example.com/link/token-1" {
		t.Errorf("TrackingURL = %q, want the wrapped URL", got)
	}
}

// A tracking URL whose token is empty is a dead link: it fails uuid validation in
// the redirect handler and sends the reader to /bad-link instead of the article.
func TestWrapItemURLsLeavesTokenlessItemsDangling(t *testing.T) {
	views := []feeds.FeedView{feedViewWithItems("item-1")}

	got := wrapItemURLs(views, map[string]string{}, "https://api.example.com")

	if url := got[0].Items[0].TrackingURL; !strings.HasSuffix(url, "/link/") {
		t.Errorf("TrackingURL = %q, want a URL ending in /link/ for an untokenized item", url)
	}
}

func TestSendEmailWithRetrySucceedsOnTheFirstAttempt(t *testing.T) {
	emailService := &fakeEmailService{}
	sch := schedulerWith(emailService, nil)
	nl := sendableNewsletter()

	result, err := sch.sendEmailWithRetry(context.Background(), nl, "<p>issue</p>")
	if err != nil {
		t.Fatalf("sendEmailWithRetry: %v", err)
	}

	if result.ID != "send-1" {
		t.Errorf("result.ID = %q, want the first send", result.ID)
	}

	attempts := emailService.sendAttempts()
	if len(attempts) != 1 {
		t.Fatalf("made %d send attempts, want 1", len(attempts))
	}

	want := sendAttempt{
		subject:        nl.Name,
		sender:         "news@example.com",
		recipient:      nl.Email,
		body:           "<p>issue</p>",
		idempotencyKey: attempts[0].idempotencyKey,
	}
	if attempts[0] != want {
		t.Errorf("send attempt = %+v, want %+v", attempts[0], want)
	}
}

func TestSendEmailWithRetryRecoversFromTransientFailures(t *testing.T) {
	emailService := &fakeEmailService{
		errs: []error{errors.New("502 bad gateway"), errors.New("connection reset")},
	}
	sch := schedulerWith(emailService, nil)

	result, err := sch.sendEmailWithRetry(context.Background(), sendableNewsletter(), "<p>issue</p>")
	if err != nil {
		t.Fatalf("sendEmailWithRetry: %v", err)
	}

	if result.ID != "send-3" {
		t.Errorf("result.ID = %q, want the third send", result.ID)
	}

	if got := len(emailService.sendAttempts()); got != 3 {
		t.Errorf("made %d send attempts, want 3", got)
	}
}

// The error has to reach the caller, because that is what triggers deleting the
// issue row that was already written for this send.
func TestSendEmailWithRetryGivesUpAfterMaxAttempts(t *testing.T) {
	lastErr := errors.New("still failing")
	errs := make([]error, maxSendAttempts)
	for i := range errs {
		errs[i] = errors.New("earlier failure")
	}
	errs[maxSendAttempts-1] = lastErr

	emailService := &fakeEmailService{errs: errs}
	sch := schedulerWith(emailService, nil)

	result, err := sch.sendEmailWithRetry(context.Background(), sendableNewsletter(), "<p>issue</p>")

	if !errors.Is(err, lastErr) {
		t.Errorf("sendEmailWithRetry error = %v, want the final attempt's error", err)
	}

	if result != nil {
		t.Errorf("sendEmailWithRetry returned %+v alongside an error", result)
	}

	if got := len(emailService.sendAttempts()); got != maxSendAttempts {
		t.Errorf("made %d send attempts, want %d", got, maxSendAttempts)
	}
}

// A retry after a timeout that actually delivered must not produce a second
// email, which only holds while every attempt carries the same key.
func TestSendEmailWithRetryReusesOneIdempotencyKey(t *testing.T) {
	emailService := &fakeEmailService{
		errs: []error{errors.New("timeout"), errors.New("timeout")},
	}
	sch := schedulerWith(emailService, nil)
	nl := sendableNewsletter()

	if _, err := sch.sendEmailWithRetry(context.Background(), nl, "<p>issue</p>"); err != nil {
		t.Fatalf("sendEmailWithRetry: %v", err)
	}

	attempts := emailService.sendAttempts()
	if len(attempts) != 3 {
		t.Fatalf("made %d send attempts, want 3", len(attempts))
	}

	firstKey := attempts[0].idempotencyKey
	if firstKey == "" {
		t.Fatal("sent with an empty idempotency key")
	}

	for i, attempt := range attempts[1:] {
		if attempt.idempotencyKey != firstKey {
			t.Errorf("attempt %d used key %q, want %q", i+2, attempt.idempotencyKey, firstKey)
		}
	}
}

// The key is keyed on the send window, so next week's issue has to be a different
// send rather than a duplicate of this one.
func TestSendEmailWithRetryKeysOnTheSendWindow(t *testing.T) {
	keyFor := func(t *testing.T, nextSendTime time.Time) string {
		t.Helper()

		emailService := &fakeEmailService{}
		sch := schedulerWith(emailService, nil)
		nl := sendableNewsletter()
		nl.NextSendTime = nextSendTime

		if _, err := sch.sendEmailWithRetry(context.Background(), nl, "<p>issue</p>"); err != nil {
			t.Fatalf("sendEmailWithRetry: %v", err)
		}

		return emailService.sendAttempts()[0].idempotencyKey
	}

	thisWeek := time.Date(2026, time.March, 10, 9, 30, 0, 0, time.UTC)
	sameWindow := keyFor(t, thisWeek)
	repeatedWindow := keyFor(t, thisWeek)
	nextWindow := keyFor(t, thisWeek.AddDate(0, 0, 7))

	if sameWindow != repeatedWindow {
		t.Errorf("the same send window produced keys %q and %q", sameWindow, repeatedWindow)
	}

	if sameWindow == nextWindow {
		t.Errorf("a later send window reused key %q", sameWindow)
	}
}

// Shutdown cancels the job context mid-backoff. Sending anyway would deliver an
// issue whose send time is never recorded, because the update never runs.
func TestSendEmailWithRetryStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emailService := &fakeEmailService{
		errs:   []error{errors.New("timeout"), errors.New("timeout")},
		onSend: func(int) { cancel() },
	}
	sch := schedulerWith(emailService, nil)
	sch.schedulerConfig.SendRetryDelay = time.Hour

	result, err := sch.sendEmailWithRetry(ctx, sendableNewsletter(), "<p>issue</p>")

	if !errors.Is(err, context.Canceled) {
		t.Errorf("sendEmailWithRetry error = %v, want context.Canceled", err)
	}

	if result != nil {
		t.Errorf("sendEmailWithRetry returned %+v after cancellation", result)
	}

	if got := len(emailService.sendAttempts()); got != 1 {
		t.Errorf("made %d send attempts after cancellation, want 1", got)
	}
}

type fetchCall struct {
	feed   feeds.BaseFeed
	since  time.Time
	userId string
}

type feedFetchOutcome struct {
	view  *feeds.FeedView
	err   error
	delay time.Duration
}

type fakeFeedFetcher struct {
	mu    sync.Mutex
	calls []fetchCall

	outcomes map[string]feedFetchOutcome
}

func (f *fakeFeedFetcher) GetFeedDataSince(
	ctx context.Context,
	feed feeds.BaseFeed,
	since time.Time,
	userId string,
) (*feeds.FeedView, error) {
	f.mu.Lock()
	f.calls = append(f.calls, fetchCall{feed, since, userId})
	f.mu.Unlock()

	outcome, ok := f.outcomes[feed.URL]
	if !ok {
		return nil, fmt.Errorf("test fetcher has no outcome scripted for %q", feed.URL)
	}

	if outcome.delay > 0 {
		time.Sleep(outcome.delay)
	}

	if outcome.err != nil {
		return nil, outcome.err
	}

	return outcome.view, nil
}

func (f *fakeFeedFetcher) fetchCalls() []fetchCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]fetchCall(nil), f.calls...)
}

func viewTitled(title string, itemIDs ...string) *feeds.FeedView {
	view := feedViewWithItems(itemIDs...)
	view.Title = title

	return &view
}

func feedAt(url string) feeds.BaseFeed {
	return feeds.BaseFeed{GlobalFeedId: url, URL: url, Name: url}
}

// The three buckets drive different sections of the issue, and an empty
// Succeeded with an empty Failed is what decides whether the issue is sent at all.
func TestFetchFeedsForNewsletterPartitionsResults(t *testing.T) {
	fetcher := &fakeFeedFetcher{
		outcomes: map[string]feedFetchOutcome{
			"https://a.example.com": {view: viewTitled("A", "item-1")},
			"https://b.example.com": {view: viewTitled("B")},
			"https://c.example.com": {view: viewTitled("C", "item-2", "item-3")},
			"https://d.example.com": {err: &feeds.FetchError{Message: "404 Not Found"}},
			"https://e.example.com": {err: errors.New("dial tcp: i/o timeout")},
		},
	}

	sch := schedulerWith(nil, fetcher)
	nl := sendableNewsletter()
	nl.Feeds = []feeds.BaseFeed{
		feedAt("https://a.example.com"),
		feedAt("https://b.example.com"),
		feedAt("https://c.example.com"),
		feedAt("https://d.example.com"),
		feedAt("https://e.example.com"),
	}

	got, err := sch.fetchFeedsForNewsletter(context.Background(), nl)
	if err != nil {
		t.Fatalf("fetchFeedsForNewsletter: %v", err)
	}

	if len(got.Succeeded) != 2 {
		t.Errorf("Succeeded has %d feeds, want the 2 with items", len(got.Succeeded))
	}

	if len(got.SucceededNoItems) != 1 {
		t.Errorf("SucceededNoItems has %d feeds, want the 1 that fetched but was empty", len(got.SucceededNoItems))
	}

	if len(got.Failed) != 2 {
		t.Fatalf("Failed has %d feeds, want 2", len(got.Failed))
	}

	// The reason is printed in the failed-feeds section of the issue, so a fetch
	// error's own message has to survive and anything else has to be generic.
	wantReasons := map[string]string{
		"https://d.example.com": "404 Not Found",
		"https://e.example.com": "Could not be retrieved",
	}
	for _, failed := range got.Failed {
		if want := wantReasons[failed.Feed.URL]; failed.Reason != want {
			t.Errorf("reason for %s = %q, want %q", failed.Feed.URL, failed.Reason, want)
		}
	}
}

// Feeds are fetched concurrently, but the reader sees them in the order they were
// added to the newsletter rather than in whatever order the network answered.
func TestFetchFeedsForNewsletterKeepsTheNewsletterOrder(t *testing.T) {
	fetcher := &fakeFeedFetcher{
		outcomes: map[string]feedFetchOutcome{
			"https://slow.example.com":   {view: viewTitled("Slow", "item-1"), delay: 60 * time.Millisecond},
			"https://medium.example.com": {view: viewTitled("Medium", "item-2"), delay: 30 * time.Millisecond},
			"https://fast.example.com":   {view: viewTitled("Fast", "item-3")},
		},
	}

	sch := schedulerWith(nil, fetcher)
	nl := sendableNewsletter()
	nl.Feeds = []feeds.BaseFeed{
		feedAt("https://slow.example.com"),
		feedAt("https://medium.example.com"),
		feedAt("https://fast.example.com"),
	}

	got, err := sch.fetchFeedsForNewsletter(context.Background(), nl)
	if err != nil {
		t.Fatalf("fetchFeedsForNewsletter: %v", err)
	}

	var titles []string
	for _, view := range got.Succeeded {
		titles = append(titles, view.Title)
	}

	want := []string{"Slow", "Medium", "Fast"}
	if strings.Join(titles, ",") != strings.Join(want, ",") {
		t.Errorf("Succeeded order = %v, want %v", titles, want)
	}
}

// Every feed is asked for items since the newsletter's last send, so a feed
// fetched with the wrong cutoff would repeat or skip a reader's items.
func TestFetchFeedsForNewsletterFetchesEachFeedOnceSinceTheLastSend(t *testing.T) {
	fetcher := &fakeFeedFetcher{
		outcomes: map[string]feedFetchOutcome{
			"https://a.example.com": {view: viewTitled("A", "item-1")},
			"https://b.example.com": {view: viewTitled("B", "item-2")},
		},
	}

	sch := schedulerWith(nil, fetcher)
	nl := sendableNewsletter()
	nl.Feeds = []feeds.BaseFeed{feedAt("https://a.example.com"), feedAt("https://b.example.com")}

	if _, err := sch.fetchFeedsForNewsletter(context.Background(), nl); err != nil {
		t.Fatalf("fetchFeedsForNewsletter: %v", err)
	}

	calls := fetcher.fetchCalls()
	if len(calls) != len(nl.Feeds) {
		t.Fatalf("made %d fetches for %d feeds", len(calls), len(nl.Feeds))
	}

	seen := make(map[string]bool, len(calls))
	for _, call := range calls {
		if seen[call.feed.URL] {
			t.Errorf("fetched %s more than once", call.feed.URL)
		}
		seen[call.feed.URL] = true

		if !call.since.Equal(nl.LastSendTime) {
			t.Errorf("fetched %s since %s, want %s", call.feed.URL, call.since, nl.LastSendTime)
		}

		if call.userId != nl.UserID {
			t.Errorf("fetched %s as user %q, want %q", call.feed.URL, call.userId, nl.UserID)
		}
	}
}

func TestFetchFeedsForNewsletterWithNoFeeds(t *testing.T) {
	sch := schedulerWith(nil, &fakeFeedFetcher{})

	got, err := sch.fetchFeedsForNewsletter(context.Background(), sendableNewsletter())
	if err != nil {
		t.Fatalf("fetchFeedsForNewsletter: %v", err)
	}

	if len(got.Succeeded) != 0 || len(got.SucceededNoItems) != 0 || len(got.Failed) != 0 {
		t.Errorf("a newsletter with no feeds produced %+v", got)
	}
}

func TestShouldSkipEmptySend(t *testing.T) {
	withItems := newsletterFetchFeedResult{Succeeded: []feeds.FeedView{feedViewWithItems("item-1")}}
	empty := newsletterFetchFeedResult{}

	tests := []struct {
		name          string
		isOneOffSend  bool
		sendWhenEmpty bool
		results       newsletterFetchFeedResult
		want          bool
	}{
		{name: "nothing was found", results: empty, want: true},
		{name: "a feed had items", results: withItems, want: false},

		// A feed that fetched cleanly but had nothing new is still an empty issue:
		// only failures and real items are worth sending on their own.
		{
			name:    "every feed fetched but none had items",
			results: newsletterFetchFeedResult{SucceededNoItems: []feeds.FeedView{feedViewWithItems()}},
			want:    true,
		},
		{
			name:    "a feed failed",
			results: newsletterFetchFeedResult{Failed: []failedFeed{{Feed: feedAt("https://a.example.com")}}},
			want:    false,
		},
		{name: "the reader asked for empty issues", sendWhenEmpty: true, results: empty, want: false},
		{name: "a one-off send was requested", isOneOffSend: true, results: empty, want: false},
		{name: "a one-off send that found items", isOneOffSend: true, results: withItems, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nl := sendableNewsletter()
			nl.IsOneOffSend = tt.isOneOffSend
			nl.SendWhenEmpty = tt.sendWhenEmpty

			if got := shouldSkipEmptySend(nl, tt.results); got != tt.want {
				t.Errorf("shouldSkipEmptySend() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Every link in the rendered issue is built here, and a wrong one is only ever
// discovered by a reader who cannot unsubscribe.
func TestAssembleNewsletterPassesTheIssueLinks(t *testing.T) {
	emailService := &fakeEmailService{assembled: "<html>issue</html>"}
	sch := schedulerWith(emailService, nil)
	nl := sendableNewsletter()

	results := newsletterFetchFeedResult{
		Succeeded:        []feeds.FeedView{feedViewWithItems("item-1")},
		SucceededNoItems: []feeds.FeedView{feedViewWithItems(), feedViewWithItems()},
		Failed:           []failedFeed{{Feed: feedAt("https://a.example.com"), Reason: "404 Not Found"}},
	}

	html, err := sch.assembleNewsletter(nl, results, "issue-9")
	if err != nil {
		t.Fatalf("assembleNewsletter: %v", err)
	}

	if html != "<html>issue</html>" {
		t.Errorf("assembleNewsletter returned %q", html)
	}

	if emailService.assembleTemplate != "newsletter.html" {
		t.Errorf("rendered template %q, want newsletter.html", emailService.assembleTemplate)
	}

	wantArgs := map[string]any{
		"NewsletterName":       "Weekly Digest",
		"UnsubscribeURL":       "https://web.example.com/unsubscribe?unsubscribeToken=unsub-token",
		"ManageNewslettersURL": "https://web.example.com/newsletters",
		"IssueURL":             "https://web.example.com/issues/issue-9",
	}
	for key, want := range wantArgs {
		if got := emailService.assembleArgs[key]; got != want {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}

	wantBucketSizes := map[string]int{"Feeds": 1, "EmptyFeeds": 2, "FailedFeeds": 1}
	for key, want := range wantBucketSizes {
		switch bucket := emailService.assembleArgs[key].(type) {
		case []feeds.FeedView:
			if len(bucket) != want {
				t.Errorf("%s has %d entries, want %d", key, len(bucket), want)
			}
		case []failedFeed:
			if len(bucket) != want {
				t.Errorf("%s has %d entries, want %d", key, len(bucket), want)
			}
		default:
			t.Errorf("%s has unexpected type %T", key, bucket)
		}
	}
}
