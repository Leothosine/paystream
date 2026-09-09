package refund

import (
	"errors"
	"testing"
)

func settledPayout() Payout {
	return Payout{ID: "bat_1", Amount: 100, Status: PayoutStatusSettled}
}

func TestRequest_succeedsOnSettledPayout(t *testing.T) {
	s := NewService()

	r, err := s.Request("ref_1", settledPayout(), 40)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusPending {
		t.Fatalf("expected StatusPending, got %s", r.Status)
	}
	if r.Amount != 40 {
		t.Fatalf("expected amount 40, got %v", r.Amount)
	}
}

func TestRequest_rejectsUnsettledPayout(t *testing.T) {
	s := NewService()
	p := Payout{ID: "bat_1", Amount: 100, Status: "pending"}

	_, err := s.Request("ref_1", p, 40)
	if !errors.Is(err, ErrPayoutNotSettled) {
		t.Fatalf("expected ErrPayoutNotSettled, got %v", err)
	}
}

func TestRequest_rejectsNonPositiveAmount(t *testing.T) {
	s := NewService()

	_, err := s.Request("ref_1", settledPayout(), 0)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}

	_, err = s.Request("ref_2", settledPayout(), -5)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount for negative amount, got %v", err)
	}
}

func TestRequest_rejectsOverRefund(t *testing.T) {
	s := NewService()
	p := settledPayout()

	if _, err := s.Request("ref_1", p, 60); err != nil {
		t.Fatalf("unexpected error on first refund: %v", err)
	}

	_, err := s.Request("ref_2", p, 50)
	if !errors.Is(err, ErrAmountExceedsRefundable) {
		t.Fatalf("expected ErrAmountExceedsRefundable, got %v", err)
	}
}

func TestRequest_allowsPartialThenRemainingRefund(t *testing.T) {
	s := NewService()
	p := settledPayout()

	if _, err := s.Request("ref_1", p, 60); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Refundable(p) != 40 {
		t.Fatalf("expected 40 refundable, got %v", s.Refundable(p))
	}

	if _, err := s.Request("ref_2", p, 40); err != nil {
		t.Fatalf("unexpected error refunding remaining balance: %v", err)
	}
	if s.Refundable(p) != 0 {
		t.Fatalf("expected 0 refundable after full refund, got %v", s.Refundable(p))
	}
}

func TestProcess_marksRefundProcessed(t *testing.T) {
	s := NewService()
	r, _ := s.Request("ref_1", settledPayout(), 25)

	processed, err := s.Process(r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed.Status != StatusProcessed {
		t.Fatalf("expected StatusProcessed, got %s", processed.Status)
	}
}

func TestProcess_unknownRefundReturnsError(t *testing.T) {
	s := NewService()

	if _, err := s.Process("does-not-exist"); err == nil {
		t.Fatalf("expected an error for unknown refund ID")
	}
}
