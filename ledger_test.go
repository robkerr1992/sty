package sty

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type memoryLedger struct {
	claimed map[OperationKey]bool
}

var _ Ledger = (*memoryLedger)(nil)

func (l *memoryLedger) Begin(_ context.Context, key OperationKey, _ Metadata) (Claim, error) {
	if l.claimed == nil {
		l.claimed = make(map[OperationKey]bool)
	}
	if l.claimed[key] {
		return Claim{}, fmt.Errorf("memory ledger: %s: %w", key, ErrAlreadyClaimed)
	}
	l.claimed[key] = true
	return Claim{Key: key, OperationID: string(key), StartedAt: time.Now().UTC(), Attempt: 1}, nil
}

func (*memoryLedger) Settle(context.Context, Claim, TerminalOutcome) error    { return nil }
func (*memoryLedger) ReapStale(context.Context, time.Duration) (int64, error) { return 0, nil }

type pullableMemoryLedger struct {
	memoryLedger
	claimNext Claim
	nextErr   error
}

var _ PullableLedger = (*pullableMemoryLedger)(nil)

func (l *pullableMemoryLedger) ClaimNext(context.Context, Metadata) (Claim, error) {
	if l.nextErr != nil {
		return Claim{}, l.nextErr
	}
	return l.claimNext, nil
}

func TestPullableLedgerClaimNext(t *testing.T) {
	want := Claim{Key: "seed-2", Attempt: 2, Feedback: "revise"}
	ledger := &pullableMemoryLedger{claimNext: want}
	got, err := ledger.ClaimNext(context.Background(), nil)
	if err != nil || got != want {
		t.Fatalf("ClaimNext() = %#v, %v; want %#v, nil", got, err, want)
	}
}

func TestErrNoPendingWorkWraps(t *testing.T) {
	err := fmt.Errorf("memory ledger: %w", ErrNoPendingWork)
	if !errors.Is(err, ErrNoPendingWork) {
		t.Fatalf("errors.Is(%v, ErrNoPendingWork) = false", err)
	}
}
func TestLedgerFreshClaim(t *testing.T) {
	ledger := &memoryLedger{}
	claim, err := ledger.Begin(context.Background(), "seed-1", nil)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	if claim.Attempt != 1 {
		t.Fatalf("Claim.Attempt = %d, want 1", claim.Attempt)
	}
}

func TestErrAlreadyClaimedWraps(t *testing.T) {
	err := fmt.Errorf("omen ledger: seed %s: %w", "seed-1", ErrAlreadyClaimed)
	if !errors.Is(err, ErrAlreadyClaimed) {
		t.Fatalf("errors.Is(%v, ErrAlreadyClaimed) = false", err)
	}
}
