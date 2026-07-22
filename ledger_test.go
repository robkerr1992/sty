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
