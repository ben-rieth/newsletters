package newsletters

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

func exportWithFeedCounts(counts ...int) NewslettersExport {
	export := NewslettersExport{Version: ExportVersion}
	for _, count := range counts {
		export.Newsletters = append(export.Newsletters, ExportableNewsletter{
			Frequency:    string(db.FrequencyDaily),
			SendTimezone: "UTC",
			Feeds:        make([]feeds.ExportableFeed, count),
		})
	}
	return export
}

func TestValidateExportCapsTotalFeedsAcrossNewsletters(t *testing.T) {
	if err := validateExport(exportWithFeedCounts(250, 250)); err != nil {
		t.Fatalf("500 feeds should be accepted, got %v", err)
	}

	err := validateExport(exportWithFeedCounts(250, 251))
	if !errors.Is(err, ErrInvalidImport) {
		t.Fatalf("501 feeds should be rejected as ErrInvalidImport, got %v", err)
	}
}

func TestValidateScheduleRejectsSchedulesTheAppCannotProduce(t *testing.T) {
	cases := []struct {
		name  string
		nl    ExportableNewsletter
		valid bool
	}{
		{"weekly saturday", ExportableNewsletter{Frequency: "weekly", SendDay: 6, SendTimezone: "UTC"}, true},
		{"weekly day 7", ExportableNewsletter{Frequency: "weekly", SendDay: 7, SendTimezone: "UTC"}, false},
		{"monthly day 0", ExportableNewsletter{Frequency: "monthly", SendDay: 0, SendTimezone: "UTC"}, false},
		{"monthly day 31", ExportableNewsletter{Frequency: "monthly", SendDay: 31, SendTimezone: "UTC"}, true},
		{"daily ignores send day", ExportableNewsletter{Frequency: "daily", SendDay: 31, SendTimezone: "UTC"}, true},
		{"empty timezone", ExportableNewsletter{Frequency: "daily", SendTimezone: ""}, false},
	}

	for _, tc := range cases {
		err := validateSchedule(tc.nl)
		if tc.valid && err != nil {
			t.Errorf("%s: expected valid, got %v", tc.name, err)
		}
		if !tc.valid && !errors.Is(err, ErrInvalidImport) {
			t.Errorf("%s: expected ErrInvalidImport, got %v", tc.name, err)
		}
	}
}

func TestNormalizeLegacySendDaysOnlyTouchesVersion1(t *testing.T) {
	legacy := NewslettersExport{Version: 1, Newsletters: []ExportableNewsletter{
		{Frequency: "weekly", SendDay: 9},
		{Frequency: "monthly", SendDay: 0},
	}}
	normalizeLegacySendDays(&legacy)
	if legacy.Newsletters[0].SendDay != 2 || legacy.Newsletters[1].SendDay != 1 {
		t.Errorf("v1 send days not normalised: %+v", legacy.Newsletters)
	}

	current := NewslettersExport{Version: 2, Newsletters: []ExportableNewsletter{
		{Frequency: "weekly", SendDay: 9},
	}}
	normalizeLegacySendDays(&current)
	if current.Newsletters[0].SendDay != 9 {
		t.Errorf("v2 send day should be left alone, got %d", current.Newsletters[0].SendDay)
	}
}

func TestStatusOrActiveDefaultsV1FilesToActive(t *testing.T) {
	cases := map[string]db.NewsletterStatus{
		"":         db.NewsletterStatusActive,
		"active":   db.NewsletterStatusActive,
		"inactive": db.NewsletterStatusInactive,
	}

	for in, want := range cases {
		if got := statusOrActive(in); got != want {
			t.Errorf("statusOrActive(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildFeedImportsKeepsAliasStatusAndFilters(t *testing.T) {
	exported := []feeds.ExportableFeed{{
		ID:     "old-id",
		URL:    "https://example.com/feed.xml",
		Alias:  "Example",
		Status: "inactive",
		Filters: []feeds.ExportableFilter{{
			Id:       "old-filter-id",
			Field:    db.FilterFieldTitle,
			Operator: db.FilterOperatorDoesNotContain,
			Pattern:  "sponsored",
		}},
	}}

	params, err := buildFeedImports("nl-id", "user-id", exported)
	if err != nil {
		t.Fatalf("buildFeedImports: %v", err)
	}
	if len(params) != 1 {
		t.Fatalf("got %d params, want 1", len(params))
	}

	p := params[0]
	if p.NewsletterID != "nl-id" || p.UserID != "user-id" || p.Url != exported[0].URL || p.Alias != "Example" {
		t.Errorf("unexpected params: %+v", p)
	}
	if p.Status != db.NewsletterStatusInactive {
		t.Errorf("status = %q, want inactive", p.Status)
	}

	var filters []feeds.ExportableFilter
	if err := json.Unmarshal(p.Filters, &filters); err != nil {
		t.Fatalf("filters are not valid JSON: %v", err)
	}
	want := exported[0].Filters
	if !reflect.DeepEqual(filters, want) {
		t.Errorf("filters = %+v, want %+v", filters, want)
	}
}

func TestBuildFeedImportsEncodesMissingFiltersAsEmptyArray(t *testing.T) {
	params, err := buildFeedImports("nl-id", "user-id", []feeds.ExportableFeed{{URL: "https://example.com"}})
	if err != nil {
		t.Fatalf("buildFeedImports: %v", err)
	}

	if string(params[0].Filters) != "[]" {
		t.Errorf("filters = %s, want []", params[0].Filters)
	}
}
