package sty

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type workerLedger struct {
	recordingLedger
	mu       sync.Mutex
	claims   []Claim
	nextErrs []error
	polls    int
	polled   chan time.Time
}

func (l *workerLedger) ClaimNext(context.Context, Metadata) (Claim, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.polls++
	if l.polled != nil {
		select {
		case l.polled <- time.Now():
		default:
		}
	}
	if len(l.nextErrs) > 0 {
		err := l.nextErrs[0]
		l.nextErrs = l.nextErrs[1:]
		return Claim{}, err
	}
	if len(l.claims) > 0 {
		claim := l.claims[0]
		l.claims = l.claims[1:]
		return claim, nil
	}
	return Claim{}, ErrNoPendingWork
}

func (l *workerLedger) pollCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.polls
}

func TestWorkerPollsEmptyQueueUntilContextEnds(t *testing.T) {
	ledger := &workerLedger{}
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()
	worker := Worker{Runner: &Runner{Ledger: ledger}, IdlePoll: 5 * time.Millisecond}
	if err := worker.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() error = %v, want deadline exceeded", err)
	}
	if polls := ledger.pollCount(); polls < 2 {
		t.Fatalf("polls = %d, want at least 2", polls)
	}
}

func TestWorkerReportsRunErrorAndContinues(t *testing.T) {
	ledger := &workerLedger{claims: []Claim{{Key: "one", Attempt: 1}}}
	reported := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	worker := Worker{
		Runner:     &Runner{Ledger: ledger, Core: coreFunc(func(context.Context, *RunState) error { return errors.New("boom") })},
		IdlePoll:   5 * time.Millisecond,
		OnRunError: func(err error) { reported <- err },
	}
	_ = worker.Run(ctx)
	select {
	case err := <-reported:
		if err.Error() != "boom" {
			t.Fatalf("reported error = %v", err)
		}
	default:
		t.Fatal("OnRunError was not called")
	}
	if polls := ledger.pollCount(); polls < 2 {
		t.Fatalf("polls = %d, want loop to continue", polls)
	}
}

func TestWorkerCancellationInterruptsSleep(t *testing.T) {
	ledger := &workerLedger{polled: make(chan time.Time, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	worker := Worker{Runner: &Runner{Ledger: ledger}, IdlePoll: time.Second}
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	<-ledger.polled
	start := time.Now()
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want canceled", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("cancellation took %v", elapsed)
	}
}

func TestWorkerSuccessfulRunPollsAgainWithoutIdleWait(t *testing.T) {
	ledger := &workerLedger{
		claims: []Claim{{Key: "one", Attempt: 1}},
		polled: make(chan time.Time, 4),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker := Worker{
		Runner:   &Runner{Ledger: ledger, Core: coreFunc(func(context.Context, *RunState) error { return nil })},
		IdlePoll: time.Second,
	}
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	first := <-ledger.polled
	second := <-ledger.polled
	cancel()
	<-done
	if gap := second.Sub(first); gap > 100*time.Millisecond {
		t.Fatalf("gap between successful run and next poll = %v, want immediate", gap)
	}
}
