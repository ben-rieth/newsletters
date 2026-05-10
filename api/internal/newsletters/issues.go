package newsletters

import (
	"context"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

type IssuesService struct {
	queries *db.Queries
}

func NewIssuesService(queries *db.Queries) *IssuesService {
	return &IssuesService{queries}
}

func (s *IssuesService) GetIssue(ctx context.Context, issueID, userID string) (*DetailedIssue, error) {
	issue, err := s.queries.GetIssue(ctx, db.GetIssueParams{
		ID:     issueID,
		UserID: userID,
	})

	if err != nil {
		return nil, err
	}

	items, err := s.queries.GetIssueItems(ctx, db.GetIssueItemsParams{
		IssueID: issueID,
		UserID:  userID,
	})

	if err != nil {
		return nil, err
	}

	issueFeeds, err := s.queries.GetIssueFeeds(ctx, db.GetIssueFeedsParams{
		IssueID: issueID,
		UserID:  userID,
	})

	if err != nil {
		return nil, err
	}

	itemsByFeed := make(map[string][]feeds.FeedItemView)
	for _, item := range items {
		itemView := feeds.FeedItemView{
			ItemID:      item.ItemID,
			Title:       item.Title,
			URL:         item.Url,
			PublishDate: item.PublishDate,
		}

		itemsByFeed[item.FeedID] = append(itemsByFeed[item.FeedID], itemView)
	}

	feedViews := make([]feeds.FeedView, 0, len(issueFeeds))
	for _, feed := range issueFeeds {
		feedViews = append(feedViews, feeds.FeedView{
			Title:   feed.Title,
			HtmlURL: feed.HtmlUrl,
			Items:   itemsByFeed[feed.ID],
		})
	}

	return &DetailedIssue{
		IssueID:        issueID,
		NewsletterName: issue.Name,
		SentAt:         issue.SentAt,
		Feeds:          feedViews,
	}, nil
}
