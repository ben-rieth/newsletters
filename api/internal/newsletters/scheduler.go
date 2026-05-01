package newsletters

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/config"
	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/email"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/google/uuid"
)

type Scheduler struct {
	newsletterService *NewsletterService
	feedService       *feeds.FeedService
	emailService      email.EmailService
	cfg               *config.Config
	schedulerConfig   *SchedulerConfig
}

type SchedulerConfig struct {
	MaxWorkers int
}

func NewScheduler(
	newsletterService *NewsletterService,
	feedService *feeds.FeedService,
	emailService email.EmailService,
	cfg *config.Config,
	schedulerConfig *SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		newsletterService,
		feedService,
		emailService,
		cfg,
		schedulerConfig,
	}
}

func (sch *Scheduler) KickOff(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		sch.pollNewslettersWithContext()

		for {
			select {
			case <-ticker.C:
				sch.pollNewslettersWithContext()
			case <-ctx.Done():
				log.Println("Shutting down newsletter scheduler")
				return
			}
		}
	}()
}

func (sch *Scheduler) ForcePoll() {
	sch.pollNewslettersWithContext()
}

func (sch *Scheduler) pollNewslettersWithContext() {
	tickCtx := context.Background()
	tickCtx, wl := wideLog.CreateWideLogAndAddToContext(tickCtx)

	tickId := uuid.New()
	wl.AddLogField("tickId", tickId)

	startTime := time.Now()

	err := sch.pollNewsletters(tickCtx)
	if err != nil {
		wl.AddErrorField(err)
	}

	endTime := time.Now()
	wl.AddLogField("duration", endTime.Sub(startTime))

	shouldKeep, level := shouldKeepSchedulerLog(wl)

	if shouldKeep {
		wl.Slog(tickCtx, level)
	}
}

func (sch *Scheduler) pollNewsletters(ctx context.Context) error {
	if err := db.WaitForDB(ctx, sch.newsletterService.db); err != nil {
		return err
	}

	dueNewsletters, err := sch.newsletterService.GetDueNewsletters(ctx)
	if err != nil {
		return err
	}

	count := len(*dueNewsletters)
	wideLog.AddLogField(ctx, "nlCount", count)
	if count == 0 {
		return nil
	}

	sem := make(chan struct{}, sch.schedulerConfig.MaxWorkers)

	var nlWaitGroup sync.WaitGroup

	for i, nl := range *dueNewsletters {
		nlWaitGroup.Add(1)
		go func(ctx context.Context, index int, nl SendableNewsletter, sem chan struct{}) {
			defer nlWaitGroup.Done()

			nlCtx, nlLog := wideLog.CreateWideLogAndAddToContext(ctx)

			startTime := time.Now()
			err := sch.buildAndSendNewsletter(nlCtx, nl, sem)
			endTime := time.Now()
			nlLog.AddLogField("duration", endTime.Sub(startTime))

			if err != nil {
				nlLog.AddErrorField(err)
			}

			shouldKeep, level := shouldKeepSchedulerLog(nlLog)
			if shouldKeep {
				nlLog.Slog(ctx, level)
			}
		}(ctx, i, nl, sem)
	}

	nlWaitGroup.Wait()
	return nil
}

func (sch *Scheduler) buildAndSendNewsletter(ctx context.Context, nl SendableNewsletter, sem chan struct{}) error {
	wideLog.AddLogField(ctx, "newsletterId", nl.ID)
	wideLog.AddLogField(ctx, "feedCount", len(nl.Feeds))

	feedResults, err := sch.fetchFeedsForNewsletter(ctx, &nl, sem)
	if err != nil {
		return err
	}

	wideLog.AddLogField(ctx, "nonEmptyFeedCount", len(feedResults.Succeeded))
	wideLog.AddLogField(ctx, "emptyFeedCount", len(feedResults.SucceededNoItems))

	newsletterHtml, err := sch.assembleNewsletter(&nl, feedResults)
	if err != nil {
		return err
	}

	result, err := sch.emailService.Send(
		ctx,
		nl.Name,
		sch.cfg.NewsletterSenderEmail.Address,
		nl.Email,
		newsletterHtml,
	)

	if err != nil {
		return err
	}
	wideLog.AddLogField(ctx, "emailSendId", result.ID)
	wideLog.AddLogField(ctx, "emailSentAt", result.Time)

	sentAt := result.Time

	err = sch.newsletterService.UpdateSendTimes(ctx, &nl, sentAt)
	if err != nil {
		return err
	}
	return nil
}

type newsletterFetchFeedResult struct {
	Succeeded        []feeds.FeedView
	SucceededNoItems []feeds.FeedView
	Failed           []feeds.BaseFeed
}

func (sch *Scheduler) fetchFeedsForNewsletter(
	ctx context.Context,
	nl *SendableNewsletter,
	sem chan struct{},
) (newsletterFetchFeedResult, error) {
	var wg sync.WaitGroup

	results := make([]*feeds.FeedView, len(nl.Feeds))
	feedErrors := make([]error, len(nl.Feeds))

	for i, feed := range nl.Feeds {
		wg.Add(1)
		sem <- struct{}{}

		go func(index int, feed feeds.BaseFeed) {
			defer wg.Done()
			defer func() { <-sem }()

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

	var failed []feeds.BaseFeed
	for i, err := range feedErrors {
		if err != nil {
			failed = append(failed, nl.Feeds[i])
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
) (string, error) {
	return sch.emailService.AssembleEmail("newsletter.html", map[string]any{
		"NewsletterName": nl.Name,
		"Feeds":          fetchResults.Succeeded,
		"EmptyFeeds":     fetchResults.SucceededNoItems,
		"FailedFeeds":    fetchResults.Failed,
		"UnsubscribeURL": fmt.Sprintf("%s/unsubscribe?unsubscribeToken=%s", sch.cfg.WebURL, nl.UnsubscribeToken),
	})
}

func shouldKeepSchedulerLog(wl *wideLog.WideLog) (bool, slog.Level) {
	if wl.HasError() {
		return true, slog.LevelError
	}

	return rand.Float64() < 0.05, slog.LevelInfo
}
