package newsletters

import (
	"fmt"
	"time"

	"github.com/ben-rieth/newsletter-api/internal/db"
	"github.com/ben-rieth/newsletter-api/internal/feeds"
)

type SendableNewsletter struct {
	ID string
	Name string
	Frequency string
	SendDay int
	SendHour int
	SendMinute int
	SendTimezone string
	LastSendTime time.Time
	Email string
	UserID string
	Feeds []feeds.BaseFeed
}

func ComputeNextSendTime(
	frequency db.Frequency, 
	sendDay, sendHour, sendMinute int, 
	sendTimezone string, 
	base time.Time,
) (time.Time, error) {
	location, err := time.LoadLocation(sendTimezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid timezone %q: %w", sendTimezone, err)
	}

	localBase := base.In(location)

	switch frequency {
	case db.FrequencyMonthly:
		clampedSendDay := clampSendDayToLastDayOfMonth(localBase.Year(), localBase.Month(), sendDay)
		candidate := time.Date(localBase.Year(), localBase.Month(), clampedSendDay, sendHour, sendMinute, 0, 0, location)

		if candidate.Before(localBase) {
			nextMonth := localBase.Month() + 1
			nextYear := localBase.Year()

			if nextMonth > 12 {
				nextMonth = 1
				nextYear++
			}

			clampedSendDay = clampSendDayToLastDayOfMonth(nextYear, nextMonth, sendDay)
			candidate = time.Date(nextYear, nextMonth, clampedSendDay, sendHour, sendMinute, 0, 0, location)
		}

		return candidate.UTC(), nil
	case db.FrequencyWeekly:
		candidate := time.Date(localBase.Year(), localBase.Month(), localBase.Day(), sendHour, sendMinute, 0, 0, location)

		daysUntil := (sendDay - int(candidate.Weekday()) + 7) % 7
		candidate = candidate.AddDate(0, 0, daysUntil)
		if candidate.Before(localBase) {
			candidate = candidate.AddDate(0, 0, 7)
		}

		return candidate.UTC(), nil
	case db.FrequencyDaily:
		candidate := time.Date(localBase.Year(), localBase.Month(), localBase.Day(), sendHour, sendMinute, 0, 0, location)

		if candidate.Before(localBase) {
			candidate = candidate.AddDate(0, 0, 1)
		}

		return candidate.UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("Unknown schedule frequency %q", frequency)
	}
}

func ComputeLastSendTime(
	frequency db.Frequency, 
	sendDay, sendHour, sendMinute int, 
	sendTimezone string, 
	base time.Time,
) (time.Time, error) {
	location, err := time.LoadLocation(sendTimezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid timezone %q: %w", sendTimezone, err)
	}

	localBase := base.In(location)

	switch frequency {
	case db.FrequencyMonthly:
		clampedSendDay := clampSendDayToLastDayOfMonth(localBase.Year(), localBase.Month(), sendDay)
		candidate := time.Date(localBase.Year(), localBase.Month(), clampedSendDay, sendHour, sendMinute, 0, 0, location)

		if candidate.After(localBase) {
			prevMonth := localBase.Month() - 1
			prevYear := localBase.Year()

			if prevMonth < 1 {
				prevMonth = 12
				prevYear--
			}

			clampedSendDay = clampSendDayToLastDayOfMonth(prevYear, prevMonth, sendDay)
			candidate = time.Date(prevYear, prevMonth, clampedSendDay, sendHour, sendMinute, 0, 0, location)
		}

		return candidate.UTC(), nil
	case db.FrequencyWeekly:
		candidate := time.Date(localBase.Year(), localBase.Month(), localBase.Day(), sendHour, sendMinute, 0, 0, location)

		daysUntil := (sendDay - int(candidate.Weekday()) + 7) % 7
		candidate = candidate.AddDate(0, 0, daysUntil)
		if candidate.After(localBase) {
			candidate = candidate.AddDate(0, 0, -7)
		}

		return candidate.UTC(), nil
	case db.FrequencyDaily:
		candidate := time.Date(localBase.Year(), localBase.Month(), localBase.Day(), sendHour, sendMinute, 0, 0, location)

		if candidate.After(localBase) {
			candidate = candidate.AddDate(0, 0, -1)
		}

		return candidate.UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("Unknown schedule frequency %q", frequency)
	}
}

func clampSendDayToLastDayOfMonth(year int, month time.Month, day int) int {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		return lastDay
	}

	return day
}