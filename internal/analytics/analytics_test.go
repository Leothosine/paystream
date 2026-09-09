package analytics

import (
	"testing"
	"time"
)

func TestSummarize_countsAndSuccessRate(t *testing.T) {
	day := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	payouts := []Payout{
		{Asset: "USDC", Amount: 100, Status: StatusSettled, CreatedAt: day},
		{Asset: "USDC", Amount: 50, Status: StatusSettled, CreatedAt: day},
		{Asset: "XLM", Amount: 25, Status: StatusFailed, CreatedAt: day},
		{Asset: "USDC", Amount: 10, Status: StatusPending, CreatedAt: day},
	}

	s := Summarize(payouts)

	if s.TotalCount != 4 {
		t.Fatalf("expected TotalCount 4, got %d", s.TotalCount)
	}
	if s.SettledCount != 2 {
		t.Fatalf("expected SettledCount 2, got %d", s.SettledCount)
	}
	if s.FailedCount != 1 {
		t.Fatalf("expected FailedCount 1, got %d", s.FailedCount)
	}
	if s.SuccessRate != 2.0/3.0 {
		t.Fatalf("expected SuccessRate 2/3, got %v", s.SuccessRate)
	}
	if s.VolumeByAsset["USDC"] != 150 {
		t.Fatalf("expected USDC volume 150, got %v", s.VolumeByAsset["USDC"])
	}
	if _, failedCounted := s.VolumeByAsset["XLM"]; failedCounted {
		t.Fatalf("failed payouts should not contribute to settled volume")
	}
}

func TestSummarize_dailyVolumeIsSortedAndGroupedByUTCDay(t *testing.T) {
	day1 := time.Date(2026, 9, 1, 23, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 9, 2, 1, 0, 0, 0, time.UTC)

	payouts := []Payout{
		{Asset: "USDC", Amount: 100, Status: StatusSettled, CreatedAt: day2},
		{Asset: "USDC", Amount: 40, Status: StatusSettled, CreatedAt: day1},
		{Asset: "USDC", Amount: 60, Status: StatusSettled, CreatedAt: day1},
	}

	s := Summarize(payouts)

	if len(s.DailyVolume) != 2 {
		t.Fatalf("expected 2 daily totals, got %d", len(s.DailyVolume))
	}
	if s.DailyVolume[0].Date != "2026-09-01" || s.DailyVolume[0].Volume != 100 {
		t.Fatalf("unexpected first day total: %+v", s.DailyVolume[0])
	}
	if s.DailyVolume[1].Date != "2026-09-02" || s.DailyVolume[1].Volume != 100 {
		t.Fatalf("unexpected second day total: %+v", s.DailyVolume[1])
	}
}

func TestSummarize_zeroResolvedGivesZeroSuccessRate(t *testing.T) {
	s := Summarize([]Payout{{Status: StatusPending}})

	if s.SuccessRate != 0 {
		t.Fatalf("expected SuccessRate 0 when nothing has resolved, got %v", s.SuccessRate)
	}
}
