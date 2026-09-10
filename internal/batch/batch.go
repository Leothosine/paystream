// Package batch processes multiple payment items submitted together as a
// single logical request, mirroring the payout batch chunking described in
// the architecture doc without depending on the Stellar transaction layer.
package batch

import (
	"errors"
	"fmt"
)

// MaxItems is the largest number of payment items PayStream will accept in a
// single batch, matching the Stellar multi-operation transaction limit.
const MaxItems = 100

// ItemStatus is the outcome of processing a single item within a batch.
type ItemStatus string

const (
	ItemStatusAccepted ItemStatus = "accepted"
	ItemStatusRejected ItemStatus = "rejected"
)

var (
	// ErrEmptyBatch is returned when a batch has no items.
	ErrEmptyBatch = errors.New("batch: at least one item is required")
	// ErrTooManyItems is returned when a batch exceeds MaxItems.
	ErrTooManyItems = errors.New("batch: exceeds maximum item count")
	// ErrDuplicateIdempotencyKey is returned when a batch is submitted twice
	// with the same idempotency key.
	ErrDuplicateIdempotencyKey = errors.New("batch: idempotency key already used")
)

// Item is a single payment within a batch request.
type Item struct {
	RecipientID string
	Amount      float64
}

// ItemResult is the outcome of validating a single Item.
type ItemResult struct {
	Item   Item
	Status ItemStatus
	Reason string
}

// Result is the outcome of processing a Batch.
type Result struct {
	ID      string
	Items   []ItemResult
	Total   float64
	Rejects int
}

// Service validates and processes payment batches, deduping repeated
// submissions by idempotency key.
type Service struct {
	seen map[string]Result
}

// NewService returns an empty batch Service.
func NewService() *Service {
	return &Service{seen: make(map[string]Result)}
}

// Process validates every item in a batch and returns a per-item Result. A
// batch that has already been processed under the same idempotencyKey
// returns the original Result rather than reprocessing.
func (s *Service) Process(idempotencyKey, batchID string, items []Item) (Result, error) {
	if existing, ok := s.seen[idempotencyKey]; ok {
		if existing.ID != batchID {
			return Result{}, fmt.Errorf("%w: %q", ErrDuplicateIdempotencyKey, idempotencyKey)
		}
		return existing, nil
	}

	if len(items) == 0 {
		return Result{}, ErrEmptyBatch
	}
	if len(items) > MaxItems {
		return Result{}, fmt.Errorf("%w: %d items, max %d", ErrTooManyItems, len(items), MaxItems)
	}

	res := Result{ID: batchID, Items: make([]ItemResult, 0, len(items))}
	for _, it := range items {
		ir := ItemResult{Item: it, Status: ItemStatusAccepted}
		switch {
		case it.RecipientID == "":
			ir.Status = ItemStatusRejected
			ir.Reason = "recipient ID is required"
		case it.Amount <= 0:
			ir.Status = ItemStatusRejected
			ir.Reason = "amount must be greater than 0"
		default:
			res.Total += it.Amount
		}
		if ir.Status == ItemStatusRejected {
			res.Rejects++
		}
		res.Items = append(res.Items, ir)
	}

	if idempotencyKey != "" {
		s.seen[idempotencyKey] = res
	}
	return res, nil
}
