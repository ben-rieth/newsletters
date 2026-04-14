package newsletters

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

type emailService interface {
	Send (ctx context.Context, subject, sender, recipient, body string) (*SendResult, error)
}

type Scheduler struct {
	newsletterService *NewsletterService
	feedService *feeds.FeedService
	emailService emailService
	schedulerConfig *SchedulerConfig
	tmpl *template.Template
}

type SchedulerConfig struct {
	MaxWorkers int
}

func NewScheduler(
	newsletterService *NewsletterService,
	feedService *feeds.FeedService,
	emailService emailService,
	schedulerConfig *SchedulerConfig,
	tmpl *template.Template,
) *Scheduler {
	return &Scheduler{
		newsletterService, 
		feedService, 
		emailService,
		schedulerConfig,
		tmpl,
	}
}

func (sch *Scheduler) KickOff(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		log.Println("Newsletter scheduler started")

		sch.pollNewsletters(ctx)

		for {
			select {
			case <-ticker.C:
				sch.pollNewsletters(ctx)
			case <- ctx.Done():
				log.Println("Shutting down newsletter scheduler")
				return
			}
		}
	}();
}

func (sch *Scheduler) ForcePoll(ctx context.Context) {
	sch.pollNewsletters(ctx)
}

func (sch *Scheduler) pollNewsletters(ctx context.Context) {
	dueNewsletters, err := sch.newsletterService.GetDueNewsletters(ctx)
	if err != nil {
		log.Printf("Error fetching due newsletters: %v", err)
		return
	}

	if len(*dueNewsletters) == 0 {
		log.Printf("No newsletters to send")
		return
	}

	sem := make(chan struct{}, sch.schedulerConfig.MaxWorkers)

	nlErrors := make([]error, len(*dueNewsletters))

	var nlWaitGroup sync.WaitGroup

	for i, nl := range *dueNewsletters {
		nlWaitGroup.Add(1)
		go func (ctx context.Context, index int, nl SendableNewsletter, sem chan struct{}) {
			defer nlWaitGroup.Done()
			feedResults, err := sch.fetchFeedsForNewsletter(ctx, &nl, sem)
			if err != nil {
				nlErrors[index] = err
				return
			}

			newsletterHtml, err := sch.assembleNewsletter(&nl, feedResults)
			if err != nil {
				nlErrors[index] = err
				return
			}

			result, err := sch.emailService.Send(
				ctx,
				nl.Name,
				"",
				nl.Email,
				newsletterHtml,
			)

			if err != nil {
				nlErrors[index] = err
				return
			}

			sentAt := result.Time

			err = sch.newsletterService.UpdateSendTimes(ctx, &nl, sentAt)
			if err != nil {
				nlErrors[index] = err
			}
		}(ctx, i, nl, sem)
	}

	nlWaitGroup.Wait()
}

type newsletterFetchFeedResult struct {
	Succeeded []feeds.FeedView
	Failed []feeds.BaseFeed
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
			defer func() { <- sem }()

			feedResult, err := sch.feedService.GetFeedDataSince(ctx, feed, nl.LastSendTime)
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
			log.Printf("Failed to fetch feed with id %s and url %s: %v", nl.Feeds[i].Id, nl.Feeds[i].URL, err)
			failed = append(failed, nl.Feeds[i])
		}
	}

	var notNilResults []feeds.FeedView
	for _, result := range results {
		if result != nil {
			notNilResults = append(notNilResults, *result)
		}
	}

	return newsletterFetchFeedResult{
		Succeeded: notNilResults,
		Failed: failed,
	}, nil
}

func (sch *Scheduler) assembleNewsletter(
	newsletter *SendableNewsletter,
	fetchResults newsletterFetchFeedResult,
) (string, error) {
	buffer := new(bytes.Buffer)
	err := sch.tmpl.ExecuteTemplate(buffer, "newsletter.html", map[string]any{
		"NewsletterName": newsletter.Name,
		"Feeds": fetchResults.Succeeded,
		"FailedFeeds": fetchResults.Failed,
		"UnsubscribeURL": "www.example.com",
	})

	if err != nil {
		return "", err
	}

	htmlString := buffer.String()
	return htmlString, nil
}