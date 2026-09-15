package feeds

import (
	"testing"
	"time"
)

func TestDisableDurationFor(t *testing.T) {
	tests := []struct {
		name          string
		priorDisables int32
		want          time.Duration
	}{
		{name: "first pause uses the base duration", priorDisables: 0, want: baseDisableDuration},
		{name: "each prior pause doubles", priorDisables: 1, want: 2 * baseDisableDuration},
		{name: "third pause", priorDisables: 2, want: 4 * baseDisableDuration},
		{name: "backoff caps at the max", priorDisables: maxBackoffDoublings, want: maxDisableDuration},
		{name: "past the doubling ceiling stays capped", priorDisables: 1000, want: maxDisableDuration},
		{name: "a negative count falls back to the base duration", priorDisables: -1, want: baseDisableDuration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := disableDurationFor(tt.priorDisables); got != tt.want {
				t.Errorf("disableDurationFor(%d) = %s, want %s", tt.priorDisables, got, tt.want)
			}
		})
	}
}

func TestDisableDurationNeverExceedsTheMax(t *testing.T) {
	for priorDisables := int32(-5); priorDisables < 64; priorDisables++ {
		got := disableDurationFor(priorDisables)

		if got <= 0 {
			t.Fatalf("disableDurationFor(%d) = %s, want a positive duration", priorDisables, got)
		}

		if got > maxDisableDuration {
			t.Errorf("disableDurationFor(%d) = %s, past the %s max", priorDisables, got, maxDisableDuration)
		}
	}
}
