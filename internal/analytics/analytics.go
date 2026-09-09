// Package analytics computes payment metrics and trends from payout records
// for the analytics dashboard.
package analytics

import (
	"sort"
	"time"
)

// PayoutStatus mirrors the batch status lifecycle described in the
// architecture doc (pending, signing, submitted, settled, failed).
type PayoutStatus string

const (
	StatusPending   PayoutStatus = "pending"
	StatusSigning   PayoutStatus = "signing"
	StatusSubmitted PayoutStatus = "submitted"
	StatusSettled   PayoutStatus = "settled"
	StatusFailed    PayoutStatus = "failed"
)

// Payout is the subset of a payout record needed for analytics.
type Payout struct {
	Asset     string
	Amount    float64
	Status    PayoutStatus
	CreatedAt time.Time
}

// Summary is the aggregate metrics returned by Summarize.
type Summary struct {
	TotalCount     int
	SettledCount   int
	FailedCount    int
	SuccessRate    float64 // SettledCount / (SettledCount + FailedCount), 0 if none resolved
	VolumeByAsset  map[string]float64
	DailyVolume    []DailyTotal
}

// DailyTotal is the settled volume for a single UTC calendar day.
type DailyTotal struct {
	Date   string // YYYY-MM-DD
	Volume float64
}

// Summarize aggregates payouts into dashboard-ready metrics.
func Summarize(payouts []Payout) Summary {
	s := Summary{
		VolumeByAsset: make(map[string]float64),
	}

	dailyTotals := make(map[string]float64)

	for _, p := range payouts {
		s.TotalCount++

		switch p.Status {
		case StatusSettled:
			s.SettledCount++
			s.VolumeByAsset[p.Asset] += p.Amount
			day := p.CreatedAt.UTC().Format("2006-01-02")
			dailyTotals[day] += p.Amount
		case StatusFailed:
			s.FailedCount++
		}
	}

	resolved := s.SettledCount + s.FailedCount
	if resolved > 0 {
		s.SuccessRate = float64(s.SettledCount) / float64(resolved)
	}

	for day, vol := range dailyTotals {
		s.DailyVolume = append(s.DailyVolume, DailyTotal{Date: day, Volume: vol})
	}
	sort.Slice(s.DailyVolume, func(i, j int) bool {
		return s.DailyVolume[i].Date < s.DailyVolume[j].Date
	})

	return s
}
