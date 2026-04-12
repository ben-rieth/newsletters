package newsletters

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/feeds"
	"github.com/mmcdole/gofeed"
)

type emailService interface {
	Send (ctx context.Context, subject, sender, recipient, body string) (*SendResult, error)
}

type Scheduler struct {
	newsletterService *NewsletterService
	rssService *feeds.RssService
	emailService emailService
	schedulerConfig *SchedulerConfig
	tmpl *template.Template
}

type SchedulerConfig struct {
	MaxWorkers int
}

func NewScheduler(
	newsletterService *NewsletterService,
	rssService *feeds.RssService,
	emailService emailService,
	schedulerConfig *SchedulerConfig,
	tmpl *template.Template,
) *Scheduler {
	return &Scheduler{
		newsletterService, 
		rssService, 
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
			rssFeeds, err := sch.fetchFeedsForNewsletter(ctx, &nl, sem)
			if err != nil {
				nlErrors[index] = err
				return
			}

			var sentAt time.Time
			if len(rssFeeds) > 0 {
				newsletterHtml, err := sch.assembleNewsletter(&nl, rssFeeds)
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

				sentAt = result.Time
			} else {
				sentAt = time.Now()
			}

			err = sch.newsletterService.UpdateSendTimes(ctx, &nl, sentAt)
		}(ctx, i, nl, sem)
	}

	nlWaitGroup.Wait()
}
  
func (sch *Scheduler) fetchFeedsForNewsletter(
	ctx context.Context,
	nl *SendableNewsletter, 
	sem chan struct{},
) ([]*gofeed.Feed, error) {
	var wg sync.WaitGroup

	rssFeeds := make([]*gofeed.Feed, len(nl.Feeds))
	errors := make([]error, len(nl.Feeds))

	for i, feed := range nl.Feeds {
		wg.Add(1)
		sem <- struct{}{}

		go func(index int, feed feeds.BaseFeed) {
			defer wg.Done()
			defer func() { <- sem }()

			rssFeed, err := sch.rssService.FetchFeed(ctx, feed.URL, feed.LastRetrievedAt)
			if err != nil {
				errors[index] = err
				rssFeeds[index] = nil
				return
			}

			rssFeeds[index] = rssFeed
			errors[index] = nil
		}(i, feed)
	}

	wg.Wait()

	var failed []string
	for i, err := range errors {
		if err != nil {
			log.Printf("Failed to fetch feed with id %s and url %s: %v", nl.Feeds[i].Id, nl.Feeds[i].URL, err)
			failed = append(failed, nl.Feeds[i].Id)
		}
	}

	// TODO: save failed feeds to display to the user in the UI or send a warning

	var results []*gofeed.Feed
	for _, feed := range rssFeeds {
		if feed != nil {
			results = append(results, feed)
		}
	}

	return results, nil
}

func (sch *Scheduler) assembleNewsletter(
	newsletter *SendableNewsletter,
	rssFeeds []*gofeed.Feed,
) (string, error) {
	buffer := new(bytes.Buffer)

	feedViews := make([]feeds.FeedView, 0, len(rssFeeds))
	for _, rssFeed := range rssFeeds {
		itemViews := make([]feeds.FeedItemView, 0, len(rssFeed.Items))
		for _, item := range rssFeed.Items {
			itemViews = append(itemViews, feeds.FeedItemView{
				Title: item.Title,
				URL: item.Link,
			})
		}
		
		feedViews = append(feedViews, feeds.FeedView{
			Title: rssFeed.Title,
			Items: itemViews,
		})
	}

	err := sch.tmpl.ExecuteTemplate(buffer, "newsletter.html", map[string]any{
		"NewsletterName": newsletter.Name,
		"Feeds": feedViews,
		"UnsubscribeURL": "www.example.com",
	})

	if err != nil {
		return "", err
	}

	htmlString := buffer.String()
	return htmlString, nil
}