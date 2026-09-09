package newsletters

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	dbgen "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Scheduler struct {
	newsletterService *NewsletterService
	feedService       *feeds.FeedService
	emailService      email.EmailService
	queries           *dbgen.Queries
	cfg               *config.Config
	schedulerConfig   *SchedulerConfig
	queue             chan string
	pollMu            sync.Mutex
}

type SchedulerConfig struct {
	MaxWorkers        int
	NewsletterTimeout int
	TickTimeout       int
}

func NewScheduler(
	newsletterService *NewsletterService,
	feedService *feeds.FeedService,
	emailService email.EmailService,
	queries *dbgen.Queries,
	cfg *config.Config,
	schedulerConfig *SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		newsletterService: newsletterService,
		feedService:       feedService,
		emailService:      emailService,
		queries:           queries,
		cfg:               cfg,
		schedulerConfig:   schedulerConfig,
		queue:             make(chan string),
	}
}

func (sch *Scheduler) KickOff(ctx context.Context) {
	sch.startSchedulerWorkers(ctx)

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		sch.enqueueDueNewsletters(ctx)

		for {
			select {
			case <-ticker.C:
				sch.enqueueDueNewsletters(ctx)
			case <-ctx.Done():
				log.Println("Shutting down newsletter scheduler")
				return
			}
		}
	}()
}

func (sch *Scheduler) ForcePoll(ctx context.Context) {
	sch.enqueueDueNewsletters(ctx)
}

func (sch *Scheduler) enqueueDueNewsletters(ctx context.Context) {
	tickCtx, cancel := context.WithTimeout(ctx, time.Duration(sch.schedulerConfig.TickTimeout)*time.Second)
	defer cancel()

	tickCtx, wl := wideLog.CreateWideLogAndAddToContext(tickCtx)
	tickId := uuid.New()
	wl.AddLogField("tickId", tickId)

	// Skip rather than queue: a waiting poll would only re-send the due set the
	// running one already covered.
	if !sch.pollMu.TryLock() {
		wl.AddMessage("Poll already in progress. Skipping.")
		wl.SlogAs(tickCtx, slog.LevelWarn, "Scheduler Tick")
		return
	}
	defer sch.pollMu.Unlock()

	startTime := time.Now()

	if err := db.WaitForDB(tickCtx, sch.newsletterService.db); err != nil {
		wl.AddErrorField(err)
		wl.SlogAs(tickCtx, slog.LevelError, "Scheduler Tick")
		return
	}

	newsletterIds, err := sch.queries.GetDueNewsletters(tickCtx)
	if err != nil {
		wl.AddErrorField(err)
		wl.SlogAs(tickCtx, slog.LevelError, "Scheduler Tick")
		return
	}

	wl.AddLogField("dueNewsletterCount", len(newsletterIds))

	for i, id := range newsletterIds {
		select {
		case sch.queue <- id:
		case <-tickCtx.Done():
			wl.AddLogField("enqueuedNewsletterCount", i)
			wl.AddErrorField(tickCtx.Err())
			wl.AddLogField("duration", time.Since(startTime).String())
			wl.SlogAs(tickCtx, slog.LevelError, "Scheduler Tick")
			return
		}
	}

	wl.AddLogField("enqueuedNewsletterCount", len(newsletterIds))

	wl.AddLogField("duration", time.Since(startTime).String())
	shouldKeep, level := shouldKeepSchedulerLog(wl)
	if shouldKeep {
		wl.SlogAs(tickCtx, level, "Scheduler Tick")
	}
}

func (sch *Scheduler) startSchedulerWorkers(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < sch.schedulerConfig.MaxWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()

			workerCtx, wl := wideLog.CreateWideLogAndAddToContext(ctx)
			wl.AddLogField("workerId", workerId)
			wl.AddMessage("Scheduler worker started")
			wl.SlogAs(workerCtx, slog.LevelInfo, "Scheduler Worker")

			for {
				select {
				case newsletterID, ok := <-sch.queue:
					if !ok {
						return
					}
					sch.processJob(ctx, newsletterID, workerId)
				case <-ctx.Done():
					wl.AddMessage("Scheduler worker shutting down")
					wl.SlogAs(workerCtx, slog.LevelInfo, "Scheduler Worker")
					return
				}
			}
		}(i)
	}

	go func() {
		<-ctx.Done()
		wg.Wait()

		shutdownCtx, wl := wideLog.CreateWideLogAndAddToContext(ctx)
		wl.AddMessage("All scheduler workers shut down")
		wl.SlogAs(shutdownCtx, slog.LevelInfo, "Scheduler Worker")
	}()
}

func (sch *Scheduler) processJob(parentCtx context.Context, newsletterID string, workerId int) {
	jobCtx, cancel := context.WithTimeout(parentCtx, time.Duration(sch.schedulerConfig.NewsletterTimeout)*time.Second)
	defer cancel()

	jobCtx, wl := wideLog.CreateWideLogAndAddToContext(jobCtx)
	wl.AddLogField("newsletterId", newsletterID)
	wl.AddLogField("workerId", workerId)

	startTime := time.Now()

	err := sch.processSingleNewsletter(jobCtx, newsletterID)
	if err != nil {
		wl.AddErrorField(err)
	}

	endTime := time.Now()
	wl.AddLogField("duration", endTime.Sub(startTime).String())

	shouldKeep, level := shouldKeepSchedulerLog(wl)

	if shouldKeep {
		wl.SlogAs(jobCtx, level, "Newsletter Job")
	}
}

func (sch *Scheduler) processSingleNewsletter(ctx context.Context, newsletterId string) error {
	nl, err := sch.newsletterService.GetSendableNewsletter(ctx, newsletterId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			wideLog.AddMessage(ctx, "Newsletter not found. Maybe deleted or marked inactive.")
			return nil
		}

		return err
	}

	wideLog.AddLogField(ctx, "feedCount", len(nl.Feeds))

	feedResults, err := sch.fetchFeedsForNewsletter(ctx, nl)
	if err != nil {
		return err
	}

	wideLog.AddLogField(ctx, "nonEmptyFeedCount", len(feedResults.Succeeded))
	wideLog.AddLogField(ctx, "emptyFeedCount", len(feedResults.SucceededNoItems))

	if !nl.IsOneOffSend && len(feedResults.Succeeded) == 0 && len(feedResults.Failed) == 0 && !nl.SendWhenEmpty {
		wideLog.AddLogField(ctx, "skippedEmpty", true)
		return sch.newsletterService.SkipSend(ctx, nl)
	}

	itemIdToTokenMap := generateTokensForItems(feedResults.Succeeded)
	feedResults.Succeeded = wrapItemURLs(
		feedResults.Succeeded,
		itemIdToTokenMap,
		sch.cfg.PublicApiURL,
	)

	issueId, err := sch.newsletterService.StoreNewsletterIssue(
		ctx, nl.ID, nl.UserID, itemIdToTokenMap,
	)
	if err != nil {
		return err
	}

	newsletterHtml, err := sch.assembleNewsletter(nl, feedResults, issueId)
	if err != nil {
		sch.cleanUpAfterSendFailure(ctx, issueId, nl.UserID)
		return err
	}

	result, err := sch.sendEmailWithRetry(ctx, nl, newsletterHtml)

	if err != nil {
		sch.cleanUpAfterSendFailure(ctx, issueId, nl.UserID)
		return err
	}
	wideLog.AddLogField(ctx, "emailSendId", result.ID)
	wideLog.AddLogField(ctx, "emailSentAt", result.Time)

	sentAt := result.Time

	err = sch.newsletterService.UpdateSendTimes(ctx, nl, sentAt)
	if err != nil {
		return err
	}

	return nil
}

func (sch *Scheduler) sendEmailWithRetry(ctx context.Context, nl *SendableNewsletter, html string) (*email.SendResult, error) {

	var err error
	var result *email.SendResult

	maxAttempts := 3
	delay := time.Second * 2

	// Keyed on the send window, not the attempt, so a retry after a timeout that
	// actually delivered does not send twice.
	idempotencyKey := fmt.Sprintf("newsletter-%s-%d", nl.ID, nl.NextSendTime.Unix())

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = sch.emailService.SendIdempotent(
			ctx,
			nl.Name,
			sch.cfg.NewsletterSenderEmail.Address,
			nl.Email,
			html,
			idempotencyKey,
		)

		if err == nil {
			return result, nil
		}

		if attempt == maxAttempts {
			wideLog.AddArrayField(ctx, "sendEmailAttempts", fmt.Sprintf("Attempt %d failed: %v. Giving up.", attempt, err))
			break
		}

		wideLog.AddArrayField(ctx, "sendEmailAttempts", fmt.Sprintf("Attempt %d failed: %v. Retrying.", attempt, err))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, err
}

type failedFeed struct {
	Feed   feeds.BaseFeed
	Reason string
}

type newsletterFetchFeedResult struct {
	Succeeded        []feeds.FeedView
	SucceededNoItems []feeds.FeedView
	Failed           []failedFeed
}

func (sch *Scheduler) fetchFeedsForNewsletter(
	ctx context.Context,
	nl *SendableNewsletter,
) (newsletterFetchFeedResult, error) {
	var wg sync.WaitGroup

	results := make([]*feeds.FeedView, len(nl.Feeds))
	feedErrors := make([]error, len(nl.Feeds))

	for i, feed := range nl.Feeds {
		wg.Add(1)

		go func(index int, feed feeds.BaseFeed) {
			defer wg.Done()

			feedResult, err := sch.feedService.GetFeedDataSince(ctx, feed, nl.LastSendTime, nl.UserID)
			if err != nil {
				feedErrors[index] = err
				results[index] = nil
				return
			}

			results[index] = feedResult
			feedErrors[index] = nil
		}(i, feed)
	}

	wg.Wait()

	var failed []failedFeed
	for i, err := range feedErrors {
		if err != nil {
			failed = append(failed, failedFeed{
				Feed:   nl.Feeds[i],
				Reason: feeds.DescribeFetchError(err),
			})
			wideLog.AddErrorField(ctx, err)
		}
	}

	var notNilResults []feeds.FeedView
	var notNilNoFeeds []feeds.FeedView
	for _, result := range results {
		if result != nil {
			if len(result.Items) == 0 {
				notNilNoFeeds = append(notNilNoFeeds, *result)
			} else {
				notNilResults = append(notNilResults, *result)
			}
		}
	}

	return newsletterFetchFeedResult{
		Succeeded:        notNilResults,
		SucceededNoItems: notNilNoFeeds,
		Failed:           failed,
	}, nil
}

func (sch *Scheduler) assembleNewsletter(
	nl *SendableNewsletter,
	fetchResults newsletterFetchFeedResult,
	issueID string,
) (string, error) {
	return sch.emailService.AssembleEmail("newsletter.html", map[string]any{
		"NewsletterName":       nl.Name,
		"Feeds":                fetchResults.Succeeded,
		"EmptyFeeds":           fetchResults.SucceededNoItems,
		"FailedFeeds":          fetchResults.Failed,
		"UnsubscribeURL":       fmt.Sprintf("%s/unsubscribe?unsubscribeToken=%s", sch.cfg.WebURL, nl.UnsubscribeToken),
		"ManageNewslettersURL": fmt.Sprintf("%s/newsletters", sch.cfg.WebURL),
		"IssueURL":             fmt.Sprintf("%s/issues/%s", sch.cfg.WebURL, issueID),
	})
}

func (sch *Scheduler) cleanUpAfterSendFailure(ctx context.Context, issueId, userId string) {
	deleteErr := sch.newsletterService.DeleteIssue(ctx, issueId, userId)

	if deleteErr != nil {
		wideLog.AddErrorField(
			ctx,
			fmt.Errorf("Failed to delete issue after failed email send: %w", deleteErr),
		)
	}
}

func shouldKeepSchedulerLog(wl *wideLog.WideLog) (bool, slog.Level) {
	if wl.HasError() {
		return true, slog.LevelError
	}

	return rand.Float64() < 0.02, slog.LevelInfo
}

func generateTokensForItems(feeds []feeds.FeedView) map[string]string {
	itemIdToTokenMap := make(map[string]string)

	for _, feed := range feeds {
		for _, item := range feed.Items {
			itemIdToTokenMap[item.ItemID] = uuid.NewString()
		}
	}

	return itemIdToTokenMap
}

func wrapItemURLs(feeds []feeds.FeedView, idToTokenMap map[string]string, urlBase string) []feeds.FeedView {
	for _, feed := range feeds {
		for i := range feed.Items {
			item := &feed.Items[i]
			feed.Items[i].TrackingURL = fmt.Sprintf("%s/link/%s", urlBase, idToTokenMap[item.ItemID])
		}
	}

	return feeds
}
