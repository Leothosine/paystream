// Package refund manages refund requests against settled payouts.
package refund

import (
	"errors"
	"fmt"
)

// Status is the lifecycle state of a refund request.
type Status string

const (
	StatusPending   Status = "pending"
	StatusProcessed Status = "processed"
)

// PayoutStatus mirrors the batch status lifecycle described in the
// architecture doc. Only a "settled" payout can be refunded.
type PayoutStatus string

const PayoutStatusSettled PayoutStatus = "settled"

var (
	// ErrPayoutNotSettled is returned when a refund is requested against a
	// payout that hasn't settled yet.
	ErrPayoutNotSettled = errors.New("refund: payout is not settled")
	// ErrAmountExceedsRefundable is returned when the requested refund
	// amount exceeds what remains refundable on the payout.
	ErrAmountExceedsRefundable = errors.New("refund: amount exceeds refundable balance")
	// ErrInvalidAmount is returned for a non-positive refund amount.
	ErrInvalidAmount = errors.New("refund: amount must be greater than 0")
)

// Payout is the subset of payout data needed to validate a refund.
type Payout struct {
	ID     string
	Amount float64
	Status PayoutStatus
}

// Refund is a single refund request against a payout.
type Refund struct {
	ID        string
	PayoutID  string
	Amount    float64
	Status    Status
}

// Service tracks refunds already issued per payout so it can enforce that
// the total refunded never exceeds the original payout amount.
type Service struct {
	refunded map[string]float64 // payoutID -> total already refunded
	refunds  map[string]Refund  // refundID -> refund
}

// NewService returns an empty refund Service.
func NewService() *Service {
	return &Service{
		refunded: make(map[string]float64),
		refunds:  make(map[string]Refund),
	}
}

// Request validates and records a new refund against a settled payout.
// It returns the created Refund in StatusPending.
func (s *Service) Request(refundID string, p Payout, amount float64) (Refund, error) {
	if amount <= 0 {
		return Refund{}, ErrInvalidAmount
	}
	if p.Status != PayoutStatusSettled {
		return Refund{}, ErrPayoutNotSettled
	}

	already := s.refunded[p.ID]
	if already+amount > p.Amount {
		return Refund{}, fmt.Errorf("%w: %.2f already refunded, %.2f requested, %.2f paid",
			ErrAmountExceedsRefundable, already, amount, p.Amount)
	}

	r := Refund{ID: refundID, PayoutID: p.ID, Amount: amount, Status: StatusPending}
	s.refunds[refundID] = r
	s.refunded[p.ID] = already + amount
	return r, nil
}

// Process marks a pending refund as processed. It is idempotent: calling it
// again on an already-processed refund is a no-op.
func (s *Service) Process(refundID string) (Refund, error) {
	r, ok := s.refunds[refundID]
	if !ok {
		return Refund{}, fmt.Errorf("refund: %q not found", refundID)
	}
	r.Status = StatusProcessed
	s.refunds[refundID] = r
	return r, nil
}

// Refundable returns how much of the payout has not yet been refunded.
func (s *Service) Refundable(p Payout) float64 {
	return p.Amount - s.refunded[p.ID]
}
