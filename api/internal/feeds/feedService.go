package feeds

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedService struct {
	rssService *RssService
	queries *db.Queries
	db *pgxpool.Pool
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
			Id: result.ID,
			Title: result.Title,
			Description: result.Description,
			URL: result.Url,
		}, nil
	}
	
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.SystemError
	}

	now := time.Now()
	itemLookBack := time.Date(now.Year(), now.Month() - 2, now.Day(), 0, 0, 0, 0, now.Location())

	fetchFeedRes, err := s.rssService.FetchFeed(ctx, url, itemLookBack)
	if err != nil {
		return nil, err
	}

	var feedId string
	if returnId {
		feedId, err = s.saveFeedDetails(context.Background(), fetchFeedRes)
		if err != nil {
			return nil, utils.SystemError
		}
	} else {
		feedId = ""
		go func () {
			_, err := s.saveFeedDetails(context.Background(), fetchFeedRes)
			if err != nil {
				log.Printf("Failed to cache feed in database: %v", err)
			}
		}()
	}

	return &FeedMetaData{
		Id: feedId,
		Title: fetchFeedRes.Feed.Title,
		Description: fetchFeedRes.Feed.Description,
		URL: fetchFeedRes.Feed.Link,
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
		Title: feed.Feed.Title,
		Url: feed.FinalUrl,
		Description: feed.Feed.Description,
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

	feedItemDetails := make([]db.SaveFeedItemDetailsParams, 0, len(feed.Feed.Items))
	for _, item := range feed.Feed.Items {
		feedItemDetails = append(feedItemDetails, db.SaveFeedItemDetailsParams{
			Title: item.Title,
			Url: item.Link,
			PublishDate: *item.PublishedParsed,
			FeedID: feedId,
			RetrievedAt: feed.RetrievedAt,
		})
	}

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

func buildUrlList(feed *FetchFeedResult, feedId string) []db.SaveFeedUrlsParams {
	feedUrls := make(map[string]db.SaveFeedUrlsParams, 0)
	feedUrls[feed.FinalUrl] = buildUrlParams(feed.FinalUrl, db.FeedurlsourceCanonical, feedId)
	feedUrls[feed.OriginalUrl] =  buildUrlParams(feed.OriginalUrl, db.FeedurlsourceUserSubmitted, feedId)
	// feedUrls[feed.Feed.FeedLink] = buildUrlParams(feed.Feed.FeedLink, db.FeedurlsourceInFeedResponse, feedId)
	// feedUrls[feed.Feed.Link] = buildUrlParams(feed.Feed.Link, db.FeedurlsourceInFeedResponse, feedId)
	// for _, url := range feed.Feed.Links {
	// 	feedUrls[url] = buildUrlParams(url, db.FeedurlsourceInFeedResponse, feedId)
	// }

	finalUrls := make([]db.SaveFeedUrlsParams, 0, len(feedUrls))
	for k,v := range feedUrls {
		if k == "" {
			continue
		}

		finalUrls = append(finalUrls, v)
	}

	return finalUrls
}

func buildUrlParams(url string, source db.Feedurlsource, feedId string) db.SaveFeedUrlsParams {
	return db.SaveFeedUrlsParams{
		FeedID: feedId,
		Url: url,
		Source: source,
	}
}