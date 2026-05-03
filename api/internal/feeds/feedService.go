package feeds

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/ben-rieth/newsletter-api/internal/wideLog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmcdole/gofeed"
)

type FeedService struct {
	rssService *RssService
	queries    *db.Queries
	db         *pgxpool.Pool
}

var FeedCacheStaleError = errors.New("Feed cache is stale.")

func NewFeedService(rssService *RssService, queries *db.Queries, db *pgxpool.Pool) *FeedService {
	return &FeedService{
		rssService,
		queries,
		db,
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

	fetchFeedRes, err := s.rssService.FetchFeed(ctx, url)
	if err != nil {
		return nil, err
	}

	fetchFeedRes.Feed = pruneFeedItemsBeforeTime(fetchFeedRes.Feed, itemLookBack)

	var feedId string
	if returnId {
		feedId, err = s.saveFeedDetails(context.Background(), fetchFeedRes)
		if err != nil {
			return nil, utils.SystemError
		}
	} else {
		feedId = ""
		go func() {
			_, err := s.saveFeedDetails(context.Background(), fetchFeedRes)
			if err != nil {
				wideLog.AddErrorField(ctx, fmt.Errorf("Failed to cache feed in database: %v", err))
			}
		}()
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

	if needToFetchLiveFeed {
		logData["feedUrl"] = feed.URL
		feedResult, err := s.rssService.FetchFeed(ctx, feed.URL)
		if err != nil {
			return nil, err
		}

		logData["fetchedItems"] = len(feedResult.Feed.Items)
		feedResult.Feed = pruneFeedItemsBeforeTime(feedResult.Feed, feed.LastRetrievedAt)
		logData["prunedItems"] = len(feedResult.Feed.Items)

		if err = s.updateFeedCache(
			ctx,
			feed.GlobalFeedId,
			feedResult,
		); err != nil {
			return nil, err
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

	finalItems = make([]FeedItemView, 0, len(items))
	for _, item := range items {
		finalItems = append(finalItems, FeedItemView{
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

func (s *FeedService) saveFeedDetails(ctx context.Context, feed *FetchFeedResult) (string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

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

func buildFeedItemParamsFromGoFeed(feed *gofeed.Feed, feedId string, itemsRetrievedAt time.Time) []db.SaveFeedItemDetailsParams {
	feedItemDetails := make([]db.SaveFeedItemDetailsParams, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		feedItemDetails = append(feedItemDetails, db.SaveFeedItemDetailsParams{
			Title:       item.Title,
			Url:         item.Link,
			PublishDate: *item.PublishedParsed,
			FeedID:      feedId,
			RetrievedAt: itemsRetrievedAt,
		})
	}

	return feedItemDetails
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
