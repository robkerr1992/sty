package sty

import (
	"context"
	"errors"
	"time"
)

// OperationKey identifies the unit of work a Ledger claims. Omen uses a seed
// ID string; ergo's Phase 5 adapter will use the caller's idempotency key.
type OperationKey string

// Metadata is small, caller-defined claim-time bookkeeping. It generalizes
// ergo's current repo/task/origin arguments without widening Ledger for each
// consumer. Values are immutable by convention.
type Metadata map[string]string

// Claim is the rehydration handle returned by Begin or ClaimNext.
type Claim struct {
	Key         OperationKey
	OperationID string
	Origin      string
	StartedAt   time.Time
	Attempt     int
	// Feedback carries the prior attempt's redelivery note. It is empty on a
	// fresh, first-ever claim.
	Feedback string
}

// ErrAlreadyClaimed is returned, possibly wrapped, when a key has a claim
// currently in flight. Keys whose prior claim is terminal are re-claimable.
var ErrAlreadyClaimed = errors.New("sty: operation already claimed")

// Ledger is the lifecycle persistence boundary generalized from omen's seed
// compare-and-set and ergo's operations table. Both existing stores already
// have Begin-, Settle-, and Reap-equivalent operations without predicate or
// schema changes.
type Ledger interface {
	// Begin atomically claims key or returns ErrAlreadyClaimed (wrapped) when a
	// claim is in flight. Implementations MUST write the claim and its
	// reap-relevant timestamp atomically in the same operation; otherwise a
	// partially recorded claim could evade ReapStale permanently.
	Begin(ctx context.Context, key OperationKey, meta Metadata) (Claim, error)

	// Settle writes the terminal record for claim. It must be idempotent when
	// racing ReapStale: whichever terminalizes first wins and the loser is a
	// harmless no-op, matching ergo's existing operations-table contract.
	Settle(ctx context.Context, claim Claim, outcome TerminalOutcome) error

	// ReapStale sweeps non-terminal claims older than olderThan, semantically
	// terminalizes them as TerminalOrphaned, and returns the number swept.
	ReapStale(ctx context.Context, olderThan time.Duration) (int64, error)
}

// PullableLedger is the optional queue-like ledger capability used by Worker.
type PullableLedger interface {
	Ledger
	ClaimNext(ctx context.Context, meta Metadata) (Claim, error)
}

// ErrNoPendingWork is returned, possibly wrapped, when ClaimNext finds no
// claimable operation.
var ErrNoPendingWork = errors.New("sty: no pending work")
