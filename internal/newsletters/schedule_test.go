package newsletters

import (
	"testing"
	"time"

	db "github.com/ben-rieth/newsletter-api/internal/db/generated"
)

const nyZone = "America/New_York"

func mustLoad(t *testing.T, zone string) *time.Location {
	t.Helper()

	location, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatalf("LoadLocation(%q): %v", zone, err)
	}

	return location
}

func localTime(t *testing.T, zone string, year int, month time.Month, day, hour, minute int) time.Time {
	t.Helper()

	return time.Date(year, month, day, hour, minute, 0, 0, mustLoad(t, zone))
}

type scheduleCase struct {
	name       string
	frequency  db.Frequency
	sendDay    int
	sendHour   int
	sendMinute int
	zone       string
	base       time.Time
	want       time.Time
}

func runScheduleCases(
	t *testing.T,
	name string,
	compute func(db.Frequency, int, int, int, string, time.Time) (time.Time, error),
	cases []scheduleCase,
) {
	t.Helper()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compute(tt.frequency, tt.sendDay, tt.sendHour, tt.sendMinute, tt.zone, tt.base)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			if !got.Equal(tt.want) {
				t.Errorf("%s = %s, want %s", name, got, tt.want.UTC())
			}

			if got.Location() != time.UTC {
				t.Errorf("%s returned location %s, want UTC", name, got.Location())
			}
		})
	}
}

func TestComputeNextSendTimeDaily(t *testing.T) {
	runScheduleCases(t, "ComputeNextSendTime", ComputeNextSendTime, []scheduleCase{
		{
			name:      "send time still ahead stays on the same local day",
			frequency: db.FrequencyDaily,
			sendHour:  9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 10, 9, 30),
		},
		{
			name:      "send time already passed moves to the next local day",
			frequency: db.FrequencyDaily,
			sendHour:  9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 10, 0),
			want: localTime(t, nyZone, 2026, time.March, 11, 9, 30),
		},
		{
			name:      "base in another zone is converted before comparing",
			frequency: db.FrequencyDaily,
			sendHour:  9, sendMinute: 30, zone: nyZone,
			// 18:00 UTC is 13:00 in New York, so the 09:30 slot is already gone.
			base: time.Date(2026, time.March, 10, 18, 0, 0, 0, time.UTC),
			want: localTime(t, nyZone, 2026, time.March, 11, 9, 30),
		},
	})
}

func TestComputeNextSendTimeWeekly(t *testing.T) {
	// 2026-03-10 is a Tuesday, so weekday 2 is "today".
	runScheduleCases(t, "ComputeNextSendTime", ComputeNextSendTime, []scheduleCase{
		{
			name:      "later in the same week",
			frequency: db.FrequencyWeekly, sendDay: int(time.Thursday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 12, 9, 30),
		},
		{
			name:      "earlier weekday wraps into next week",
			frequency: db.FrequencyWeekly, sendDay: int(time.Monday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 16, 9, 30),
		},
		{
			name:      "today with the hour still ahead",
			frequency: db.FrequencyWeekly, sendDay: int(time.Tuesday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 10, 9, 30),
		},
		{
			name:      "today with the hour already passed skips a full week",
			frequency: db.FrequencyWeekly, sendDay: int(time.Tuesday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 10, 0),
			want: localTime(t, nyZone, 2026, time.March, 17, 9, 30),
		},
		{
			name:      "sunday is weekday zero",
			frequency: db.FrequencyWeekly, sendDay: int(time.Sunday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 15, 9, 30),
		},
	})
}

func TestComputeNextSendTimeMonthly(t *testing.T) {
	runScheduleCases(t, "ComputeNextSendTime", ComputeNextSendTime, []scheduleCase{
		{
			name:      "day 31 clamps to the last day of a short month",
			frequency: db.FrequencyMonthly, sendDay: 31,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.February, 1, 8, 0),
			want: localTime(t, nyZone, 2026, time.February, 28, 9, 30),
		},
		{
			name:      "day 31 clamps to february 29 in a leap year",
			frequency: db.FrequencyMonthly, sendDay: 31,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2028, time.February, 1, 8, 0),
			want: localTime(t, nyZone, 2028, time.February, 29, 9, 30),
		},
		{
			name:      "clamping is recomputed for the month rolled into",
			frequency: db.FrequencyMonthly, sendDay: 31,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			// February's send already happened, and March is long enough for day 31.
			base: localTime(t, nyZone, 2026, time.February, 28, 10, 0),
			want: localTime(t, nyZone, 2026, time.March, 31, 9, 30),
		},
		{
			name:      "december rolls the year over",
			frequency: db.FrequencyMonthly, sendDay: 15,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.December, 20, 8, 0),
			want: localTime(t, nyZone, 2027, time.January, 15, 9, 30),
		},
		{
			name:      "an unset send day means the first of the month",
			frequency: db.FrequencyMonthly, sendDay: 0,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.April, 1, 9, 30),
		},
		{
			name:      "an unset send day on the last day of a month rolls forward",
			frequency: db.FrequencyMonthly, sendDay: 0,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 31, 14, 0),
			want: localTime(t, nyZone, 2026, time.April, 1, 9, 30),
		},
	})
}

func TestComputeLastSendTime(t *testing.T) {
	runScheduleCases(t, "ComputeLastSendTime", ComputeLastSendTime, []scheduleCase{
		{
			name:      "daily send time already passed today",
			frequency: db.FrequencyDaily,
			sendHour:  9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 10, 0),
			want: localTime(t, nyZone, 2026, time.March, 10, 9, 30),
		},
		{
			name:      "daily send time not yet reached falls back a day",
			frequency: db.FrequencyDaily,
			sendHour:  9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 9, 9, 30),
		},
		{
			name:      "weekly falls back to the previous occurrence",
			frequency: db.FrequencyWeekly, sendDay: int(time.Thursday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 5, 9, 30),
		},
		{
			name:      "weekly today with the hour already passed stays today",
			frequency: db.FrequencyWeekly, sendDay: int(time.Tuesday),
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 10, 0),
			want: localTime(t, nyZone, 2026, time.March, 10, 9, 30),
		},
		{
			name:      "monthly clamps the previous month it falls back to",
			frequency: db.FrequencyMonthly, sendDay: 31,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.February, 28, 9, 30),
		},
		{
			name:      "january rolls the year back",
			frequency: db.FrequencyMonthly, sendDay: 15,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.January, 10, 8, 0),
			want: localTime(t, nyZone, 2025, time.December, 15, 9, 30),
		},
		{
			name:      "an unset send day means the first of the month",
			frequency: db.FrequencyMonthly, sendDay: 0,
			sendHour: 9, sendMinute: 30, zone: nyZone,
			base: localTime(t, nyZone, 2026, time.March, 10, 8, 0),
			want: localTime(t, nyZone, 2026, time.March, 1, 9, 30),
		},
	})
}

// A send hour that does not exist locally on a spring-forward day still has to
// resolve to a real instant rather than an error or a skipped send. The zone
// resolves the gap with the pre-transition offset, so a 02:30 send lands at
// 01:30 local on that one day.
func TestComputeNextSendTimeSpringForward(t *testing.T) {
	base := localTime(t, nyZone, 2026, time.March, 8, 1, 0)

	got, err := ComputeNextSendTime(db.FrequencyDaily, 0, 2, 30, nyZone, base)
	if err != nil {
		t.Fatalf("ComputeNextSendTime: %v", err)
	}

	if !got.After(base) {
		t.Errorf("ComputeNextSendTime = %s, want an instant after %s", got, base)
	}

	want := time.Date(2026, time.March, 8, 6, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("ComputeNextSendTime = %s, want %s", got, want)
	}
}

// A local hour that happens twice on a fall-back day has two valid instants, and
// which one the zone picks is not guaranteed. Either is a correct send; landing
// outside the pair, or before the base, is not.
func TestComputeNextSendTimeFallBack(t *testing.T) {
	base := localTime(t, nyZone, 2026, time.November, 1, 0, 30)

	got, err := ComputeNextSendTime(db.FrequencyDaily, 0, 1, 30, nyZone, base)
	if err != nil {
		t.Fatalf("ComputeNextSendTime: %v", err)
	}

	beforeShift := time.Date(2026, time.November, 1, 5, 30, 0, 0, time.UTC)
	afterShift := time.Date(2026, time.November, 1, 6, 30, 0, 0, time.UTC)

	if !got.Equal(beforeShift) && !got.Equal(afterShift) {
		t.Errorf("ComputeNextSendTime = %s, want %s or %s", got, beforeShift, afterShift)
	}

	if !got.After(base) {
		t.Errorf("ComputeNextSendTime = %s, want an instant after %s", got, base)
	}
}

// A next send time in the past leaves the newsletter due on every scheduler tick,
// so it resends until the schedule catches up. The whole input space the API
// accepts has to stay on the correct side of the base.
func TestComputeSendTimesBracketTheBase(t *testing.T) {
	bases := []time.Time{
		time.Date(2026, time.March, 10, 14, 17, 43, 0, time.UTC),
		// The last day of a month, and the last day of a short one.
		time.Date(2026, time.March, 31, 14, 17, 43, 0, time.UTC),
		time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC),
		time.Date(2026, time.December, 31, 22, 0, 0, 0, time.UTC),
		time.Date(2028, time.February, 29, 3, 0, 0, 0, time.UTC),
	}

	frequencies := []db.Frequency{db.FrequencyDaily, db.FrequencyWeekly, db.FrequencyMonthly}
	zones := []string{"UTC", nyZone, "Asia/Kolkata", "Pacific/Auckland"}

	for _, base := range bases {
		for _, frequency := range frequencies {
			for _, zone := range zones {
				// 0 is what the API stores when no send day is given, and 31 is the
				// maximum it accepts.
				for sendDay := range 32 {
					next, err := ComputeNextSendTime(frequency, sendDay, 9, 30, zone, base)
					if err != nil {
						t.Fatalf("ComputeNextSendTime(%s, %d, %s): %v", frequency, sendDay, zone, err)
					}

					last, err := ComputeLastSendTime(frequency, sendDay, 9, 30, zone, base)
					if err != nil {
						t.Fatalf("ComputeLastSendTime(%s, %d, %s): %v", frequency, sendDay, zone, err)
					}

					if next.Before(base) {
						t.Errorf("%s/%s/day %d: next send %s is before base %s", frequency, zone, sendDay, next, base)
					}

					if last.After(base) {
						t.Errorf("%s/%s/day %d: last send %s is after base %s", frequency, zone, sendDay, last, base)
					}
				}
			}
		}
	}
}

func TestComputeSendTimesRejectBadInput(t *testing.T) {
	base := time.Date(2026, time.March, 10, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		frequency db.Frequency
		zone      string
	}{
		{name: "unknown timezone", frequency: db.FrequencyDaily, zone: "Mars/Olympus_Mons"},
		{name: "unknown frequency", frequency: db.Frequency("hourly"), zone: "UTC"},
		{name: "empty frequency", frequency: db.Frequency(""), zone: "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ComputeNextSendTime(tt.frequency, 1, 9, 30, tt.zone, base); err == nil {
				t.Error("ComputeNextSendTime accepted invalid input")
			}

			if _, err := ComputeLastSendTime(tt.frequency, 1, 9, 30, tt.zone, base); err == nil {
				t.Error("ComputeLastSendTime accepted invalid input")
			}
		})
	}
}
