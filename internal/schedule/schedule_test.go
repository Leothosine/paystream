package schedule

import (
	"errors"
	"testing"
	"time"
)

func TestCreate_succeedsWithValidInput(t *testing.T) {
	s := NewService()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	sc, err := s.Create("sch_1", "rec_1", 500, IntervalMonthly, start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.Status != StatusActive {
		t.Fatalf("expected StatusActive, got %s", sc.Status)
	}
	if !sc.NextRunAt.Equal(start) {
		t.Fatalf("expected NextRunAt %v, got %v", start, sc.NextRunAt)
	}
}

func TestCreate_rejectsEmptyRecipient(t *testing.T) {
	s := NewService()
	_, err := s.Create("sch_1", "", 500, IntervalDaily, time.Now())
	if !errors.Is(err, ErrEmptyRecipient) {
		t.Fatalf("expected ErrEmptyRecipient, got %v", err)
	}
}

func TestCreate_rejectsNonPositiveAmount(t *testing.T) {
	s := NewService()
	_, err := s.Create("sch_1", "rec_1", 0, IntervalDaily, time.Now())
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreate_rejectsInvalidInterval(t *testing.T) {
	s := NewService()
	_, err := s.Create("sch_1", "rec_1", 100, Interval("yearly"), time.Now())
	if !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("expected ErrInvalidInterval, got %v", err)
	}
}

func TestAdvance_movesNextRunAtByInterval(t *testing.T) {
	s := NewService()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Create("sch_1", "rec_1", 100, IntervalWeekly, start)

	sc, err := s.Advance("sch_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := start.AddDate(0, 0, 7)
	if !sc.NextRunAt.Equal(want) {
		t.Fatalf("expected NextRunAt %v, got %v", want, sc.NextRunAt)
	}
	if sc.RunCount != 1 {
		t.Fatalf("expected RunCount 1, got %d", sc.RunCount)
	}
}

func TestAdvance_rejectsPausedSchedule(t *testing.T) {
	s := NewService()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Create("sch_1", "rec_1", 100, IntervalDaily, start)
	s.Pause("sch_1")

	_, err := s.Advance("sch_1")
	if !errors.Is(err, ErrNotActive) {
		t.Fatalf("expected ErrNotActive, got %v", err)
	}
}

func TestResume_rejectsCanceledSchedule(t *testing.T) {
	s := NewService()
	s.Create("sch_1", "rec_1", 100, IntervalDaily, time.Now())
	s.Cancel("sch_1")

	if _, err := s.Resume("sch_1"); err == nil {
		t.Fatalf("expected an error resuming a canceled schedule")
	}
}

func TestDue_reportsWhetherScheduleIsDue(t *testing.T) {
	s := NewService()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Create("sch_1", "rec_1", 100, IntervalDaily, start)

	before, err := s.Due("sch_1", start.Add(-time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if before {
		t.Fatalf("expected schedule not due before NextRunAt")
	}

	after, err := s.Due("sch_1", start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !after {
		t.Fatalf("expected schedule due at NextRunAt")
	}
}
