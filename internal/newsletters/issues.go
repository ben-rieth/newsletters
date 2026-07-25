package newsletters

import (
	"context"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IssuesService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewIssuesService(queries *db.Queries, pool *pgxpool.Pool) *IssuesService {
	return &IssuesService{queries, pool}
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

	itemsByFeed := make(map[string][]IssueItem)
	for _, item := range items {
		itemView := IssueItem{
			ItemID:      item.ItemID,
			Title:       item.Title,
			PublishDate: item.PublishDate,
			State:       item.State,
			Token:       item.Token,
		}

		itemsByFeed[item.FeedID] = append(itemsByFeed[item.FeedID], itemView)
	}

	feedViews := make([]IssueFeed, 0, len(issueFeeds))
	for _, feed := range issueFeeds {
		feedViews = append(feedViews, IssueFeed{
			Title:   feed.Title,
			HtmlURL: feed.HtmlUrl,
			Items:   itemsByFeed[feed.ID],
		})
	}

	return &DetailedIssue{
		IssueID:        issueID,
		NewsletterID:   issue.NewsletterID,
		NewsletterName: issue.Name,
		SentAt:         issue.SentAt,
		Feeds:          feedViews,
	}, nil
}
