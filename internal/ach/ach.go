// Package ach adds ACH bank transfer support alongside PayStream's Stellar
// payout rail, for recipients who receive local-currency payouts through a
// traditional bank account rather than a Stellar-connected wallet.
package ach

import (
	"errors"
	"fmt"
	"regexp"
)

// TransactionType mirrors the NACHA transaction codes PayStream supports.
type TransactionType string

const (
	// TransactionTypeCredit pushes funds into the recipient's bank account.
	TransactionTypeCredit TransactionType = "credit"
	// TransactionTypeDebit pulls funds from the recipient's bank account.
	TransactionTypeDebit TransactionType = "debit"
)

// Status is the lifecycle state of an ACH Transfer.
type Status string

const (
	StatusPending   Status = "pending"
	StatusSubmitted Status = "submitted"
	StatusSettled   Status = "settled"
	StatusReturned  Status = "returned"
)

var (
	// ErrInvalidRoutingNumber is returned when the routing number is not 9
	// digits or fails the ABA checksum.
	ErrInvalidRoutingNumber = errors.New("ach: invalid routing number")
	// ErrInvalidAccountNumber is returned for an empty or malformed account number.
	ErrInvalidAccountNumber = errors.New("ach: invalid account number")
	// ErrInvalidAmount is returned for a non-positive transfer amount.
	ErrInvalidAmount = errors.New("ach: amount must be greater than 0")
	// ErrInvalidTransactionType is returned for a type other than credit/debit.
	ErrInvalidTransactionType = errors.New("ach: unsupported transaction type")
	// ErrNotPending is returned when Submit or Return is called on a transfer
	// that has already left the pending state.
	ErrNotPending = errors.New("ach: transfer is not pending")
	// ErrNotSubmitted is returned when Settle is called on a transfer that
	// hasn't been submitted yet.
	ErrNotSubmitted = errors.New("ach: transfer has not been submitted")
)

var accountNumberPattern = regexp.MustCompile(`^[0-9]{4,17}$`)

// Transfer is a single ACH bank transfer.
type Transfer struct {
	ID              string
	RecipientID     string
	RoutingNumber   string
	AccountNumber   string
	Type            TransactionType
	Amount          float64
	Status          Status
	ReturnCode      string
}

// Service validates and tracks the lifecycle of ACH transfers.
type Service struct {
	transfers map[string]Transfer
}

// NewService returns an empty ACH Service.
func NewService() *Service {
	return &Service{transfers: make(map[string]Transfer)}
}

// Create validates transfer details and records a new Transfer in
// StatusPending.
func (s *Service) Create(id, recipientID, routingNumber, accountNumber string, txType TransactionType, amount float64) (Transfer, error) {
	if amount <= 0 {
		return Transfer{}, ErrInvalidAmount
	}
	if txType != TransactionTypeCredit && txType != TransactionTypeDebit {
		return Transfer{}, fmt.Errorf("%w: %q", ErrInvalidTransactionType, txType)
	}
	if !ValidRoutingNumber(routingNumber) {
		return Transfer{}, fmt.Errorf("%w: %q", ErrInvalidRoutingNumber, routingNumber)
	}
	if !accountNumberPattern.MatchString(accountNumber) {
		return Transfer{}, fmt.Errorf("%w: must be 4-17 digits", ErrInvalidAccountNumber)
	}

	t := Transfer{
		ID:            id,
		RecipientID:   recipientID,
		RoutingNumber: routingNumber,
		AccountNumber: accountNumber,
		Type:          txType,
		Amount:        amount,
		Status:        StatusPending,
	}
	s.transfers[id] = t
	return t, nil
}

// Submit moves a pending transfer to StatusSubmitted, meaning it has been
// handed off to the ACH network.
func (s *Service) Submit(id string) (Transfer, error) {
	t, ok := s.transfers[id]
	if !ok {
		return Transfer{}, fmt.Errorf("ach: %q not found", id)
	}
	if t.Status != StatusPending {
		return Transfer{}, ErrNotPending
	}
	t.Status = StatusSubmitted
	s.transfers[id] = t
	return t, nil
}

// Settle marks a submitted transfer as settled once the ACH network confirms
// final funds movement.
func (s *Service) Settle(id string) (Transfer, error) {
	t, ok := s.transfers[id]
	if !ok {
		return Transfer{}, fmt.Errorf("ach: %q not found", id)
	}
	if t.Status != StatusSubmitted {
		return Transfer{}, ErrNotSubmitted
	}
	t.Status = StatusSettled
	s.transfers[id] = t
	return t, nil
}

// Return marks a pending or submitted transfer as returned by the receiving
// bank, recording the NACHA return code (e.g. "R01" for insufficient funds).
func (s *Service) Return(id, returnCode string) (Transfer, error) {
	t, ok := s.transfers[id]
	if !ok {
		return Transfer{}, fmt.Errorf("ach: %q not found", id)
	}
	if t.Status != StatusPending && t.Status != StatusSubmitted {
		return Transfer{}, fmt.Errorf("ach: cannot return a transfer in status %q", t.Status)
	}
	t.Status = StatusReturned
	t.ReturnCode = returnCode
	s.transfers[id] = t
	return t, nil
}

// ValidRoutingNumber reports whether s is a 9-digit ABA routing number that
// passes the standard checksum:
// 3*(d1+d4+d7) + 7*(d2+d5+d8) + (d3+d6+d9) is a multiple of 10.
func ValidRoutingNumber(s string) bool {
	if len(s) != 9 {
		return false
	}
	digits := make([]int, 9)
	for i, r := range s {
		if r < '0' || r > '9' {
			return false
		}
		digits[i] = int(r - '0')
	}
	sum := 3*(digits[0]+digits[3]+digits[6]) +
		7*(digits[1]+digits[4]+digits[7]) +
		1*(digits[2]+digits[5]+digits[8])
	return sum%10 == 0
}
