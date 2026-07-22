package sty

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRunnerBeginFailureSkipsExecuteAndSettle(t *testing.T) {
	beginErr := errors.New("claim failed")
	ledger := &recordingLedger{beginErr: beginErr}
	executed := false
	runner := Runner{Ledger: ledger, Core: coreFunc(func(context.Context, *RunState) error { executed = true; return nil })}
	err := runner.Run(context.Background(), "key", nil)
	if !errors.Is(err, beginErr) {
		t.Fatalf("Run() error = %v, want %v", err, beginErr)
	}
	if executed {
		t.Fatal("Core.Execute called after Begin failure")
	}
	if len(ledger.outcomes) != 0 {
		t.Fatal("Settle called after Begin failure")
	}
}

func TestRunnerOrdinaryErrorReturnedUnmodifiedWhenSettleSucceeds(t *testing.T) {
	execErr := errors.New("execute")
	runner := Runner{Ledger: &recordingLedger{}, Core: coreFunc(func(context.Context, *RunState) error { return execErr })}
	got := runner.Run(context.Background(), "key", nil)
	if got != execErr {
		t.Fatalf("Run() error = %v, want exact Execute error %v", got, execErr)
	}
}

func TestRunnerRecoversPanicAndSettles(t *testing.T) {
	for _, value := range []any{"boom", errors.New("boom error")} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			ledger := &recordingLedger{}
			runner := Runner{Ledger: ledger, Core: coreFunc(func(context.Context, *RunState) error { panic(value) })}
			err := runner.Run(context.Background(), "key", nil)
			var panicErr *PanicError
			if !errors.As(err, &panicErr) {
				t.Fatalf("Run() error = %T %v, want *PanicError", err, err)
			}
			if panicErr.Source != "core" || panicErr.Value != value || len(panicErr.Stack) == 0 {
				t.Fatalf("PanicError = %#v", panicErr)
			}
			if len(ledger.outcomes) != 1 || ledger.outcomes[0].Status != TerminalPanicked {
				t.Fatalf("outcomes = %#v", ledger.outcomes)
			}
		})
	}
}

type settleContextLedger struct {
	recordingLedger
	settleDone chan struct{}
	ctxErr     error
}

func (l *settleContextLedger) Settle(ctx context.Context, claim Claim, outcome TerminalOutcome) error {
	l.ctxErr = ctx.Err()
	close(l.settleDone)
	return l.recordingLedger.Settle(ctx, claim, outcome)
}

func TestRunnerSettlesAfterCancellationAndTimeout(t *testing.T) {
	tests := []struct {
		name   string
		ctx    func() (context.Context, context.CancelFunc)
		want   error
		status TerminalStatus
	}{
		{"canceled", func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }, context.Canceled, TerminalCanceled},
		{"timed out", func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), 10*time.Millisecond)
		}, context.DeadlineExceeded, TerminalTimedOut},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()
			ledger := &settleContextLedger{settleDone: make(chan struct{})}
			started := make(chan struct{})
			runner := Runner{Ledger: ledger, Core: coreFunc(func(ctx context.Context, _ *RunState) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			})}
			if tc.want == context.Canceled {
				go func() { <-started; cancel() }()
			}
			err := runner.Run(ctx, "key", nil)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Run() error = %v, want errors.Is %v", err, tc.want)
			}
			select {
			case <-ledger.settleDone:
			default:
				t.Fatal("Settle did not complete")
			}
			if ledger.ctxErr != nil {
				t.Fatalf("Settle context already done: %v", ledger.ctxErr)
			}
			if ledger.outcomes[0].Status != tc.status {
				t.Fatalf("status = %s, want %s", ledger.outcomes[0].Status, tc.status)
			}
		})
	}
}

func TestRunnerJoinsExecuteAndSettleErrors(t *testing.T) {
	execErr := errors.New("execute")
	settleErr := errors.New("settle")
	ledger := &recordingLedger{settleErr: settleErr}
	var hookCalls int
	var hooked error
	runner := Runner{
		Ledger:        ledger,
		Core:          coreFunc(func(context.Context, *RunState) error { return execErr }),
		OnSettleError: func(err error) { hookCalls++; hooked = err },
	}
	err := runner.Run(context.Background(), "key", nil)
	if !errors.Is(err, execErr) || !errors.Is(err, settleErr) {
		t.Fatalf("Run() error = %v, want both errors", err)
	}
	if hookCalls != 1 || hooked != settleErr {
		t.Fatalf("hook calls/error = %d/%v, want 1/raw settle error", hookCalls, hooked)
	}
}

type concurrentLedger struct {
	mu     sync.Mutex
	begins int
}

func (l *concurrentLedger) Begin(_ context.Context, key OperationKey, _ Metadata) (Claim, error) {
	l.mu.Lock()
	l.begins++
	l.mu.Unlock()
	return Claim{Key: key, OperationID: string(key), StartedAt: time.Now(), Attempt: 1}, nil
}
func (*concurrentLedger) Settle(context.Context, Claim, TerminalOutcome) error    { return nil }
func (*concurrentLedger) ReapStale(context.Context, time.Duration) (int64, error) { return 0, nil }

func TestRunnerConcurrentFirstCalls(t *testing.T) {
	ledger := &concurrentLedger{}
	runner := Runner{Ledger: ledger, Core: coreFunc(func(context.Context, *RunState) error { return nil })}
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := runner.Run(context.Background(), OperationKey(fmt.Sprint(i)), nil); err != nil {
				t.Errorf("Run: %v", err)
			}
		}(i)
	}
	wg.Wait()
}
