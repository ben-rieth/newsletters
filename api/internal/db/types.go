package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}