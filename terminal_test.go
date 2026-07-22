package sty

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestTerminalEnumZeroValuesAndStrings(t *testing.T) {
	if TerminalStatus(0) != TerminalUnknown {
		t.Fatal("TerminalStatus zero value is not TerminalUnknown")
	}
	if ExternalStateDisposition(0) != DispositionUnknown {
		t.Fatal("ExternalStateDisposition zero value is not DispositionUnknown")
	}
	if got := TerminalStatus(9).String(); got != "TerminalStatus(9)" {
		t.Fatalf("TerminalStatus(9).String() = %q", got)
	}
	if got := ExternalStateDisposition(3).String(); got != "ExternalStateDisposition(3)" {
		t.Fatalf("ExternalStateDisposition(3).String() = %q", got)
	}
}

func TestRunnerNeverSelfReportsOrphaned(t *testing.T) {
	tests := []struct {
		name string
		ctx  func() (context.Context, context.CancelFunc)
		core Core
		want TerminalStatus
	}{
		{"success", backgroundContext, coreFunc(func(context.Context, *RunState) error { return nil }), TerminalAccepted},
		{"ordinary error", backgroundContext, coreFunc(func(context.Context, *RunState) error { return errors.New("no") }), TerminalRejected},
		{"cancellation", canceledContext, coreFunc(func(ctx context.Context, _ *RunState) error { return ctx.Err() }), TerminalCanceled},
		{"timeout", expiredContext, coreFunc(func(ctx context.Context, _ *RunState) error { return ctx.Err() }), TerminalTimedOut},
		{"panic", backgroundContext, coreFunc(func(context.Context, *RunState) error { panic("boom") }), TerminalPanicked},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()
			ledger := &recordingLedger{}
			runner := Runner{Ledger: ledger, Core: tc.core}
			_ = runner.Run(ctx, "key", nil)
			if len(ledger.outcomes) != 1 {
				t.Fatalf("Settle calls = %d, want 1", len(ledger.outcomes))
			}
			if got := ledger.outcomes[0].Status; got != tc.want {
				t.Fatalf("status = %s, want %s", got, tc.want)
			}
			if ledger.outcomes[0].Status == TerminalOrphaned {
				t.Fatal("Runner self-reported TerminalOrphaned")
			}
		})
	}
}

type coreFunc func(context.Context, *RunState) error

func (f coreFunc) Execute(ctx context.Context, run *RunState) error { return f(ctx, run) }

type recordingLedger struct {
	beginErr  error
	settleErr error
	outcomes  []TerminalOutcome
}

func (l *recordingLedger) Begin(_ context.Context, key OperationKey, _ Metadata) (Claim, error) {
	if l.beginErr != nil {
		return Claim{}, l.beginErr
	}
	return Claim{Key: key, OperationID: string(key), StartedAt: time.Now().UTC(), Attempt: 1}, nil
}
func (l *recordingLedger) Settle(_ context.Context, _ Claim, outcome TerminalOutcome) error {
	l.outcomes = append(l.outcomes, outcome)
	return l.settleErr
}
func (*recordingLedger) ReapStale(context.Context, time.Duration) (int64, error) { return 0, nil }

func backgroundContext() (context.Context, context.CancelFunc) {
	return context.Background(), func() {}
}
func canceledContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx, func() {}
}
func expiredContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	return ctx, cancel
}

func TestTerminalStrings(t *testing.T) {
	for value, want := range map[TerminalStatus]string{
		TerminalUnknown: "TerminalUnknown", TerminalAccepted: "TerminalAccepted",
		TerminalAcceptedWithCaveats: "TerminalAcceptedWithCaveats", TerminalRejected: "TerminalRejected",
		TerminalExhausted: "TerminalExhausted", TerminalCanceled: "TerminalCanceled",
		TerminalTimedOut: "TerminalTimedOut", TerminalPanicked: "TerminalPanicked", TerminalOrphaned: "TerminalOrphaned",
	} {
		if got := value.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", value, got, want)
		}
	}
	for value, want := range map[ExternalStateDisposition]string{
		DispositionUnknown: "DispositionUnknown", NothingLeft: "NothingLeft", ArtifactsLeft: "ArtifactsLeft",
	} {
		if got := value.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", value, got, want)
		}
	}
	_ = fmt.Sprintf("%v", TerminalUnknown)
}
