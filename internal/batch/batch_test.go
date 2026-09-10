package batch

import (
	"errors"
	"testing"
)

func TestProcess_acceptsValidItems(t *testing.T) {
	s := NewService()
	items := []Item{
		{RecipientID: "rec_1", Amount: 100},
		{RecipientID: "rec_2", Amount: 250.50},
	}

	res, err := s.Process("idem_1", "bat_1", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Rejects != 0 {
		t.Fatalf("expected 0 rejects, got %d", res.Rejects)
	}
	if res.Total != 350.50 {
		t.Fatalf("expected total 350.50, got %v", res.Total)
	}
	for _, ir := range res.Items {
		if ir.Status != ItemStatusAccepted {
			t.Fatalf("expected item %+v to be accepted, got %s", ir.Item, ir.Status)
		}
	}
}

func TestProcess_rejectsInvalidItemsWithoutFailingWholeBatch(t *testing.T) {
	s := NewService()
	items := []Item{
		{RecipientID: "rec_1", Amount: 100},
		{RecipientID: "", Amount: 50},
		{RecipientID: "rec_3", Amount: -5},
	}

	res, err := s.Process("idem_1", "bat_1", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Rejects != 2 {
		t.Fatalf("expected 2 rejects, got %d", res.Rejects)
	}
	if res.Total != 100 {
		t.Fatalf("expected total 100 (only the valid item), got %v", res.Total)
	}
}

func TestProcess_rejectsEmptyBatch(t *testing.T) {
	s := NewService()
	_, err := s.Process("idem_1", "bat_1", nil)
	if !errors.Is(err, ErrEmptyBatch) {
		t.Fatalf("expected ErrEmptyBatch, got %v", err)
	}
}

func TestProcess_rejectsTooManyItems(t *testing.T) {
	s := NewService()
	items := make([]Item, MaxItems+1)
	for i := range items {
		items[i] = Item{RecipientID: "rec", Amount: 1}
	}

	_, err := s.Process("idem_1", "bat_1", items)
	if !errors.Is(err, ErrTooManyItems) {
		t.Fatalf("expected ErrTooManyItems, got %v", err)
	}
}

func TestProcess_isIdempotentForSameKeyAndBatch(t *testing.T) {
	s := NewService()
	items := []Item{{RecipientID: "rec_1", Amount: 100}}

	first, err := s.Process("idem_1", "bat_1", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := s.Process("idem_1", "bat_1", items)
	if err != nil {
		t.Fatalf("unexpected error on replay: %v", err)
	}
	if second.ID != first.ID || second.Total != first.Total {
		t.Fatalf("expected replay to return identical result, got %+v vs %+v", first, second)
	}
}

func TestProcess_rejectsIdempotencyKeyReuseWithDifferentBatch(t *testing.T) {
	s := NewService()
	items := []Item{{RecipientID: "rec_1", Amount: 100}}

	if _, err := s.Process("idem_1", "bat_1", items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := s.Process("idem_1", "bat_2", items)
	if !errors.Is(err, ErrDuplicateIdempotencyKey) {
		t.Fatalf("expected ErrDuplicateIdempotencyKey, got %v", err)
	}
}
