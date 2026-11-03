package feeds

import (
	"context"
	"errors"
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/jobs"
	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmcdole/gofeed"
	"golang.org/x/sync/singleflight"
)

type FeedService struct {
	rssService *RssService
	queries    *db.Queries
	db         *pgxpool.Pool
	jobQueue   *jobs.JobQueue
	refreshes  singleflight.Group
}

var FeedCacheStaleError = errors.New("Feed cache is stale.")

func NewFeedService(rssService *RssService, queries *db.Queries, db *pgxpool.Pool, jobQueue *jobs.JobQueue) *FeedService {
	return &FeedService{
		rssService: rssService,
		queries:    queries,
		db:         db,
		jobQueue:   jobQueue,
	}
}

func (s *FeedService) GetFeedMetaData(ctx context.Context, url string, returnId bool) (*FeedMetaData, error) {
	result, err := s.queries.GetCachedFeedDetails(ctx, url)
	if err == nil {
		return &FeedMetaData{
			Id:          result.ID,
			Title:       result.Title,
			Description: result.Description,
			URL:         result.Url,
			HtmlURL:     result.HtmlUrl,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.SystemError
	}

	now := time.Now()
	itemLookBack := now.AddDate(0, -2, 0)

	fetchFeedRes, err := s.fetchFeed(ctx, "", url)
	if err != nil {
		return nil, err
	}

	fetchFeedRes.Feed = pruneFeedItemsBeforeTime(fetchFeedRes.Feed, itemLookBack)

	feedId, err := s.saveFeedDetails(ctx, fetchFeedRes)
	if err != nil {
		return nil, utils.SystemError
	}

	return &FeedMetaData{
		Id:          feedId,
		Title:       fetchFeedRes.Feed.Title,
		Description: fetchFeedRes.Feed.Description,
		URL:         fetchFeedRes.FinalUrl,
		HtmlURL:     fetchFeedRes.Feed.Link,
	}, nil
}

func (s *FeedService) GetFeedDataSince(
	ctx context.Context,
	feed BaseFeed,
	since time.Time,
	userId string,
) (*FeedView, error) {
	logData := make(map[string]any)
	defer wideLog.AddArrayField(ctx, "GetFeedDataSince", logData)

	logData["globalFeedId"] = feed.GlobalFeedId
	logData["newsletterFeedId"] = feed.GlobalFeedId

	oneHourAgo := time.Now().Add(time.Hour * -1)

	var finalItems []FeedItemView

	needToFetchLiveFeed := feed.LastRetrievedAt.Before(oneHourAgo)
	logData["mustFetchLiveFeed"] = needToFetchLiveFeed

	breakerOpen := false

	if needToFetchLiveFeed {
		logData["feedUrl"] = feed.URL

		// Concurrent sends sharing a feed all see the same stale last_retrieved_at,
		// so without this each one would fetch the feed and count its failures.
		result, err, shared := s.refreshes.Do(feed.GlobalFeedId, func() (any, error) {
			return s.refreshFeedCache(ctx, feed)
		})
		logData["sharedRefresh"] = shared

		// A paused feed can stay paused for weeks, so keep delivering whatever is
		// already cached instead of skipping the feed until the breaker closes —
		// last_send_time advances either way and those items would be lost.
		if errors.Is(err, ErrFeedDisabled) {
			breakerOpen = true
			logData["feedBreakerOpen"] = true
		} else if err != nil {
			return nil, err
		} else {
			counts := result.(refreshCounts)
			logData["fetchedItems"] = counts.fetched
			logData["prunedItems"] = counts.pruned
		}
	}

	items, err := s.queries.GetFeedItemsPublishedAfter(ctx, db.GetFeedItemsPublishedAfterParams{
		PublishDateGreaterThan: since,
		GlobalFeedID:           feed.GlobalFeedId,
		NewsletterFeedID:       feed.NewsletterFeedId,
		UserID:                 userId,
	})

	if err != nil {
		return nil, err
	}

	logData["itemsForNewsletter"] = len(items)

	// Nothing cached to fall back on, so surface the pause to the reader.
	if breakerOpen && len(items) == 0 {
		return nil, ErrFeedDisabled
	}

	finalItems = make([]FeedItemView, 0, len(items))
	for _, item := range items {
		finalItems = append(finalItems, FeedItemView{
			ItemID:      item.ID,
			Title:       item.Title,
			URL:         item.Url,
			PublishDate: item.PublishDate,
		})
	}

	return &FeedView{
		Items:   finalItems,
		HtmlURL: feed.HtmlURL,
		Title:   feed.Name,
	}, nil
}

type refreshCounts struct {
	fetched int
	pruned  int
}

func (s *FeedService) refreshFeedCache(ctx context.Context, feed BaseFeed) (refreshCounts, error) {
	feedResult, err := s.fetchFeed(ctx, feed.GlobalFeedId, feed.URL)
	if err != nil {
		return refreshCounts{}, err
	}

	counts := refreshCounts{fetched: len(feedResult.Feed.Items)}
	feedResult.Feed = pruneFeedItemsBeforeTime(feedResult.Feed, feed.LastRetrievedAt)
	counts.pruned = len(feedResult.Feed.Items)

	return counts, s.updateFeedCache(ctx, feed.GlobalFeedId, feedResult)
}

func (s *FeedService) saveFeedDetails(ctx context.Context, feed *FetchFeedResult) (string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	existing, err := qtx.GetFeedIdsForUrls(ctx, []string{feed.FinalUrl})
	if err != nil {
		return "", err
	}
	if len(existing) > 0 {
		feedId := existing[0].FeedID
		err = qtx.SaveFeedUrl(ctx, db.SaveFeedUrlParams{
			FeedID: feedId,
			Url:    feed.OriginalUrl,
			Source: db.FeedUrlSourceUserSubmitted,
		})
		if err != nil {
			return "", err
		}
		return feedId, tx.Commit(ctx)
	}

	feedId, err := qtx.SaveFeedDetails(ctx, db.SaveFeedDetailsParams{
		Title:           feed.Feed.Title,
		HtmlUrl:         feed.Feed.Link,
		Url:             feed.FinalUrl,
		Description:     feed.Feed.Description,
		LastRetrievedAt: feed.RetrievedAt,
	})
	if err != nil {
		return "", err
	}

	feedUrls := buildUrlList(feed, feedId)
	_, err = qtx.SaveFeedUrls(ctx, feedUrls)
	if err != nil {
		return "", err
	}

	feedItemDetails := buildFeedItemParamsFromGoFeed(feed.Feed, feedId, feed.RetrievedAt)

	_, err = qtx.SaveFeedItemDetails(ctx, feedItemDetails)
	if err != nil {
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}

	return feedId, nil
}

func (s *FeedService) updateFeedCache(
	ctx context.Context,
	feedId string,
	feedResult *FetchFeedResult,
) error {
	feedItemDetails := buildFeedItemParamsFromGoFeed(feedResult.Feed, feedId, feedResult.RetrievedAt)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	_, err = qtx.SaveFeedItemDetails(ctx, feedItemDetails)
	if err != nil {
		return err
	}

	err = qtx.UpdateCachedFeed(ctx, db.UpdateCachedFeedParams{
		ID:              feedId,
		LastRetrievedAt: feedResult.RetrievedAt,
		Title:           feedResult.Feed.Title,
		HtmlUrl:         feedResult.Feed.Link,
		Description:     feedResult.Feed.Description,
	})

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func buildFeedItemParamsFromGoFeed(feed *gofeed.Feed, feedId string, itemsRetrievedAt time.Time) db.SaveFeedItemDetailsParams {
	params := db.SaveFeedItemDetailsParams{
		FeedID:         feedId,
		RetrievedAt:    itemsRetrievedAt,
		Titles:         make([]string, 0, len(feed.Items)),
		Urls:           make([]string, 0, len(feed.Items)),
		NormalizedUrls: make([]string, 0, len(feed.Items)),
		PublishDates:   make([]time.Time, 0, len(feed.Items)),
	}

	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		params.Titles = append(params.Titles, item.Title)
		params.Urls = append(params.Urls, item.Link)
		params.NormalizedUrls = append(params.NormalizedUrls, normalizeItemUrl(item.Link))
		params.PublishDates = append(params.PublishDates, *item.PublishedParsed)
	}

	return params
}

func buildUrlList(feed *FetchFeedResult, feedId string) []db.SaveFeedUrlsParams {
	feedUrls := make(map[string]db.SaveFeedUrlsParams, 0)
	feedUrls[feed.OriginalUrl] = buildUrlParams(feed.OriginalUrl, db.FeedUrlSourceUserSubmitted, feedId)
	feedUrls[feed.FinalUrl] = buildUrlParams(feed.FinalUrl, db.FeedUrlSourceCanonical, feedId)

	finalUrls := make([]db.SaveFeedUrlsParams, 0, len(feedUrls))
	for k, v := range feedUrls {
		if k == "" {
			continue
		}

		finalUrls = append(finalUrls, v)
	}

	return finalUrls
}

func buildUrlParams(url string, source db.FeedUrlSource, feedId string) db.SaveFeedUrlsParams {
	return db.SaveFeedUrlsParams{
		FeedID: feedId,
		Url:    url,
		Source: source,
	}
}

func pruneFeedItemsBeforeTime(feed *gofeed.Feed, threshold time.Time) *gofeed.Feed {
	prunedItems := make([]*gofeed.Item, 0)
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		publishedAt := *item.PublishedParsed

		if publishedAt.Before(threshold) {
			continue
		}

		prunedItems = append(prunedItems, item)
	}

	feed.Items = prunedItems

	return feed
}
