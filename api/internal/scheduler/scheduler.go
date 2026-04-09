package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/service"
	"github.com/ben-rieth/newsletter-api/internal/types"
)

type Scheduler struct {
	newsletterService *service.NewsletterService
	rssService *service.RssService
	config *SchedulerConfig
}

type SchedulerConfig struct {
	MaxWorkers int
}

func NewScheduler(
	newsletterService *service.NewsletterService,
	rssService *service.RssService,
	config *SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		newsletterService: newsletterService, 
		rssService: rssService, 
		config: config,
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

	sem := make(chan struct{}, sch.config.MaxWorkers)

	for _, nl := range *dueNewsletters {
		sch.fetchFeedsForNewsletter(ctx, &nl, sem)
	}
}
  
func (sch *Scheduler) fetchFeedsForNewsletter(
	ctx context.Context,
	nl *types.NewsletterWithFeeds, 
	sem chan struct{},
) (*[]types.RSSFeed, error) {
	var wg sync.WaitGroup

	rssFeeds := make([]*types.RSSFeed, len(nl.Feeds))
	errors := make([]error, len(nl.Feeds))

	for i, feed := range nl.Feeds {
		wg.Add(1)
		sem <- struct{}{}

		go func(index int, feed types.BaseFeed) {
			defer wg.Done()
			defer func() { <- sem }()

			rssFeed, err := sch.rssService.FetchFeed(ctx, feed.URL)
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

	var results []types.RSSFeed
	for _, feed := range rssFeeds {
		if feed != nil {
			results = append(results, *feed)
		}
	}

	return &results, nil
}