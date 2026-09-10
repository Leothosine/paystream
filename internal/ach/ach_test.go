package ach

import (
	"errors"
	"testing"
)

const validRouting = "021000021"

func TestValidRoutingNumber(t *testing.T) {
	if !ValidRoutingNumber(validRouting) {
		t.Fatalf("expected %q to be valid", validRouting)
	}
	if ValidRoutingNumber("123456789") {
		t.Fatalf("expected 123456789 to fail the checksum")
	}
	if ValidRoutingNumber("12345") {
		t.Fatalf("expected a 5-digit string to be invalid")
	}
}

func TestCreate_succeedsWithValidInput(t *testing.T) {
	s := NewService()

	tr, err := s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionTypeCredit, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Status != StatusPending {
		t.Fatalf("expected StatusPending, got %s", tr.Status)
	}
}

func TestCreate_rejectsInvalidRoutingNumber(t *testing.T) {
	s := NewService()
	_, err := s.Create("ach_1", "rec_1", "123456789", "123456789", TransactionTypeCredit, 500)
	if !errors.Is(err, ErrInvalidRoutingNumber) {
		t.Fatalf("expected ErrInvalidRoutingNumber, got %v", err)
	}
}

func TestCreate_rejectsInvalidAccountNumber(t *testing.T) {
	s := NewService()
	_, err := s.Create("ach_1", "rec_1", validRouting, "abc", TransactionTypeCredit, 500)
	if !errors.Is(err, ErrInvalidAccountNumber) {
		t.Fatalf("expected ErrInvalidAccountNumber, got %v", err)
	}
}

func TestCreate_rejectsNonPositiveAmount(t *testing.T) {
	s := NewService()
	_, err := s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionTypeCredit, 0)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreate_rejectsInvalidTransactionType(t *testing.T) {
	s := NewService()
	_, err := s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionType("wire"), 500)
	if !errors.Is(err, ErrInvalidTransactionType) {
		t.Fatalf("expected ErrInvalidTransactionType, got %v", err)
	}
}

func TestLifecycle_pendingToSubmittedToSettled(t *testing.T) {
	s := NewService()
	s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionTypeCredit, 500)

	submitted, err := s.Submit("ach_1")
	if err != nil {
		t.Fatalf("unexpected error submitting: %v", err)
	}
	if submitted.Status != StatusSubmitted {
		t.Fatalf("expected StatusSubmitted, got %s", submitted.Status)
	}

	settled, err := s.Settle("ach_1")
	if err != nil {
		t.Fatalf("unexpected error settling: %v", err)
	}
	if settled.Status != StatusSettled {
		t.Fatalf("expected StatusSettled, got %s", settled.Status)
	}
}

func TestSettle_rejectsUnsubmittedTransfer(t *testing.T) {
	s := NewService()
	s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionTypeCredit, 500)

	_, err := s.Settle("ach_1")
	if !errors.Is(err, ErrNotSubmitted) {
		t.Fatalf("expected ErrNotSubmitted, got %v", err)
	}
}

func TestReturn_recordsReturnCodeFromPending(t *testing.T) {
	s := NewService()
	s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionTypeCredit, 500)

	tr, err := s.Return("ach_1", "R01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Status != StatusReturned {
		t.Fatalf("expected StatusReturned, got %s", tr.Status)
	}
	if tr.ReturnCode != "R01" {
		t.Fatalf("expected return code R01, got %s", tr.ReturnCode)
	}
}

func TestSubmit_rejectsAlreadySubmittedTransfer(t *testing.T) {
	s := NewService()
	s.Create("ach_1", "rec_1", validRouting, "123456789", TransactionTypeCredit, 500)
	s.Submit("ach_1")

	_, err := s.Submit("ach_1")
	if !errors.Is(err, ErrNotPending) {
		t.Fatalf("expected ErrNotPending, got %v", err)
	}
}
