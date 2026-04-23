package feeds

import (
	"context"
	"errors"
	"log"
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
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.SystemError
	}

	now := time.Now()
	itemLookBack := time.Date(now.Year(), now.Month()-2, now.Day(), 0, 0, 0, 0, now.Location())

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
				log.Printf("Failed to cache feed in database: %v", err)
			}
		}()
	}

	return &FeedMetaData{
		Id:          feedId,
		Title:       fetchFeedRes.Feed.Title,
		Description: fetchFeedRes.Feed.Description,
		URL:         fetchFeedRes.Feed.Link,
	}, nil
}

func (s *FeedService) GetFeedDataSince(
	ctx context.Context,
	feed BaseFeed,
	since time.Time,
	userId string,
) (*FeedView, error) {
	wideLog.AddLogField(ctx, "globalFeedId", feed.GlobalFeedId)
	wideLog.AddLogField(ctx, "newsletterFeedId", feed.GlobalFeedId)

	oneHourAgo := time.Now().Add(time.Hour * -1)

	var finalItems []FeedItemView

	needToFetchLiveFeed := feed.LastRetrievedAt.Before(oneHourAgo)
	wideLog.AddLogField(ctx, "mustFetchLiveFeed", needToFetchLiveFeed)

	if needToFetchLiveFeed {
		wideLog.AddLogField(ctx, "feedUrl", feed.URL)
		feedResult, err := s.rssService.FetchFeed(ctx, feed.URL)
		if err != nil {
			return nil, err
		}

		wideLog.AddLogField(ctx, "fetchedItems", feedResult.Feed.Items)
		feedResult.Feed = pruneFeedItemsBeforeTime(feedResult.Feed, feed.LastRetrievedAt)
		wideLog.AddLogField(ctx, "prunedItem", feedResult.Feed.Items)

		feedItemDetails := buildFeedItemParamsFromGoFeed(feedResult.Feed, feed.GlobalFeedId, feedResult.RetrievedAt)
		if err = s.updateFeedCache(
			ctx,
			feed.GlobalFeedId,
			feedItemDetails,
			feedResult.RetrievedAt,
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

	wideLog.AddLogField(ctx, "itemsForNewsletter", len(items))

	finalItems = make([]FeedItemView, 0, len(items))
	for _, item := range items {
		finalItems = append(finalItems, FeedItemView{
			Title:       item.Title,
			URL:         item.Url,
			PublishDate: item.PublishDate,
		})
	}

	return &FeedView{
		Items: finalItems,
		Title: feed.Name,
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
	feedItemDetails []db.SaveFeedItemDetailsParams,
	retrievedAt time.Time,
) error {
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

	err = qtx.UpdateFeedLastRetrievedTime(ctx, db.UpdateFeedLastRetrievedTimeParams{
		ID:              feedId,
		LastRetrievedAt: retrievedAt,
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

	// TODO: research if I should include links in the feed data itself
	// feedUrls[feed.Feed.FeedLink] = buildUrlParams(feed.Feed.FeedLink, db.FeedurlsourceInFeedResponse, feedId)
	// feedUrls[feed.Feed.Link] = buildUrlParams(feed.Feed.Link, db.FeedurlsourceInFeedResponse, feedId)
	// for _, url := range feed.Feed.Links {
	// 	feedUrls[url] = buildUrlParams(url, db.FeedurlsourceInFeedResponse, feedId)
	// }

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
