package domain

import "time"

// Usage represents Anthropic OAuth usage data for one account.
type Usage struct {
	FiveHour *Window `json:"fiveHour,omitempty"`
	SevenDay *Window `json:"sevenDay,omitempty"`
	FetchedAt time.Time `json:"fetchedAt"`
}

// Window is a single utilization window (5h or 7d).
type Window struct {
	UtilizationPct float64   `json:"utilizationPct"`
	ResetsAt       time.Time `json:"resetsAt"`
}

// SecondsUntilReset returns countdown in seconds (0 if past).
func (w *Window) SecondsUntilReset(now time.Time) int64 {
	if w == nil {
		return 0
	}
	diff := w.ResetsAt.Sub(now).Seconds()
	if diff < 0 {
		return 0
	}
	return int64(diff)
}
