// Package schedule manages recurring payment schedules: payouts that repeat
// at a fixed interval (daily, weekly, or monthly) rather than firing once.
package schedule

import (
	"errors"
	"fmt"
	"time"
)

// Interval is the recurrence cadence of a Schedule.
type Interval string

const (
	IntervalDaily   Interval = "daily"
	IntervalWeekly  Interval = "weekly"
	IntervalMonthly Interval = "monthly"
)

// Status is the lifecycle state of a Schedule.
type Status string

const (
	StatusActive   Status = "active"
	StatusPaused   Status = "paused"
	StatusCanceled Status = "canceled"
)

var (
	// ErrInvalidInterval is returned for an interval outside the supported set.
	ErrInvalidInterval = errors.New("schedule: unsupported interval")
	// ErrInvalidAmount is returned for a non-positive recurring amount.
	ErrInvalidAmount = errors.New("schedule: amount must be greater than 0")
	// ErrEmptyRecipient is returned when no recipient ID is supplied.
	ErrEmptyRecipient = errors.New("schedule: recipient ID is required")
	// ErrNotActive is returned when an operation requires an active schedule
	// but the schedule has been paused or canceled.
	ErrNotActive = errors.New("schedule: schedule is not active")
)

// Schedule is a single recurring payment definition.
type Schedule struct {
	ID          string
	RecipientID string
	Amount      float64
	Interval    Interval
	Status      Status
	NextRunAt   time.Time
	RunCount    int
}

// Service creates and advances recurring payment schedules.
type Service struct {
	schedules map[string]Schedule
}

// NewService returns an empty schedule Service.
func NewService() *Service {
	return &Service{schedules: make(map[string]Schedule)}
}

// Create validates and stores a new recurring schedule, due to run first at
// startAt.
func (s *Service) Create(id, recipientID string, amount float64, interval Interval, startAt time.Time) (Schedule, error) {
	if recipientID == "" {
		return Schedule{}, ErrEmptyRecipient
	}
	if amount <= 0 {
		return Schedule{}, ErrInvalidAmount
	}
	if !validInterval(interval) {
		return Schedule{}, fmt.Errorf("%w: %q", ErrInvalidInterval, interval)
	}

	sc := Schedule{
		ID:          id,
		RecipientID: recipientID,
		Amount:      amount,
		Interval:    interval,
		Status:      StatusActive,
		NextRunAt:   startAt,
	}
	s.schedules[id] = sc
	return sc, nil
}

// Pause stops a schedule from being advanced until resumed.
func (s *Service) Pause(id string) (Schedule, error) {
	sc, ok := s.schedules[id]
	if !ok {
		return Schedule{}, fmt.Errorf("schedule: %q not found", id)
	}
	sc.Status = StatusPaused
	s.schedules[id] = sc
	return sc, nil
}

// Resume reactivates a paused schedule.
func (s *Service) Resume(id string) (Schedule, error) {
	sc, ok := s.schedules[id]
	if !ok {
		return Schedule{}, fmt.Errorf("schedule: %q not found", id)
	}
	if sc.Status == StatusCanceled {
		return Schedule{}, fmt.Errorf("schedule: %q is canceled and cannot be resumed", id)
	}
	sc.Status = StatusActive
	s.schedules[id] = sc
	return sc, nil
}

// Cancel permanently stops a schedule from running again.
func (s *Service) Cancel(id string) (Schedule, error) {
	sc, ok := s.schedules[id]
	if !ok {
		return Schedule{}, fmt.Errorf("schedule: %q not found", id)
	}
	sc.Status = StatusCanceled
	s.schedules[id] = sc
	return sc, nil
}

// Advance runs a due, active schedule and computes its next run time. It
// returns ErrNotActive if the schedule is paused or canceled.
func (s *Service) Advance(id string) (Schedule, error) {
	sc, ok := s.schedules[id]
	if !ok {
		return Schedule{}, fmt.Errorf("schedule: %q not found", id)
	}
	if sc.Status != StatusActive {
		return Schedule{}, ErrNotActive
	}

	sc.RunCount++
	sc.NextRunAt = nextRun(sc.NextRunAt, sc.Interval)
	s.schedules[id] = sc
	return sc, nil
}

// Due reports whether the schedule is active and its next run time is at or
// before now.
func (s *Service) Due(id string, now time.Time) (bool, error) {
	sc, ok := s.schedules[id]
	if !ok {
		return false, fmt.Errorf("schedule: %q not found", id)
	}
	return sc.Status == StatusActive && !sc.NextRunAt.After(now), nil
}

func validInterval(i Interval) bool {
	switch i {
	case IntervalDaily, IntervalWeekly, IntervalMonthly:
		return true
	default:
		return false
	}
}

func nextRun(from time.Time, interval Interval) time.Time {
	switch interval {
	case IntervalDaily:
		return from.AddDate(0, 0, 1)
	case IntervalWeekly:
		return from.AddDate(0, 0, 7)
	case IntervalMonthly:
		return from.AddDate(0, 1, 0)
	default:
		return from
	}
}
