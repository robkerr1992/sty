package sty

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// Core is the domain work a Runner drives.
type Core interface {
	Execute(ctx context.Context, run *RunState) error
}

// RunState is intentionally minimal in Phase 3.
type RunState struct {
	Claim Claim
}

// Runner owns the fixed Begin -> Execute -> deferred Settle lifecycle.
type Runner struct {
	Ledger Ledger
	Core   Core
	Plan   Plan

	// SettleTimeout bounds settlement independently of caller cancellation.
	// Zero uses DefaultSettleTimeout. Later consumers should re-examine whether
	// the default fits their own settlement work.
	SettleTimeout time.Duration
	// OnSettleError receives the raw settlement error synchronously.
	OnSettleError func(error)

	planOnce sync.Once
	planErr  error
}

const DefaultSettleTimeout = 30 * time.Second

// ErrNotPullable is returned when BeginNext is used with a plain Ledger.
var ErrNotPullable = errors.New("sty: ledger does not implement PullableLedger (no ClaimNext)")

// RetryableError lets a Core wrap a failure with feedback intended for the
// next run attempt. It does not affect terminal-status classification.
//
// Error and Unwrap are defined on *RetryableError. A non-pointer value does
// not satisfy error and is invisible to retryFeedback; always construct it as
// &RetryableError{...}.
type RetryableError struct {
	Feedback string
	Err      error
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

func retryFeedback(execErr error) string {
	var retryable *RetryableError
	if errors.As(execErr, &retryable) {
		return retryable.Feedback
	}
	return ""
}

func (r *Runner) validatePlan() error {
	r.planOnce.Do(func() { r.planErr = r.Plan.Validate() })
	return r.planErr
}

// Begin validates the Plan once and synchronously claims key. It does not
// execute or settle, allowing callers to claim before launching a goroutine.
func (r *Runner) Begin(ctx context.Context, key OperationKey, meta Metadata) (Claim, error) {
	if err := r.validatePlan(); err != nil {
		return Claim{}, fmt.Errorf("sty: phase plan validation: %w", err)
	}
	claim, err := r.Ledger.Begin(ctx, key, meta)
	if err != nil {
		return Claim{}, fmt.Errorf("sty: begin: %w", err)
	}
	return claim, nil
}

// BeginNext validates the Plan once and claims the next pending operation from
// a PullableLedger.
func (r *Runner) BeginNext(ctx context.Context, meta Metadata) (Claim, error) {
	if err := r.validatePlan(); err != nil {
		return Claim{}, fmt.Errorf("sty: phase plan validation: %w", err)
	}
	pullable, ok := r.Ledger.(PullableLedger)
	if !ok {
		return Claim{}, fmt.Errorf("sty: begin next: %w", ErrNotPullable)
	}
	claim, err := pullable.ClaimNext(ctx, meta)
	if err != nil {
		return Claim{}, fmt.Errorf("sty: begin next: %w", err)
	}
	return claim, nil
}

// RunClaim executes and settles an already-obtained claim.
func (r *Runner) RunClaim(ctx context.Context, claim Claim) (err error) {
	var execErr error
	defer func() {
		outcome := TerminalOutcome{EndedAt: time.Now().UTC()}

		if rec := recover(); rec != nil {
			outcome.Status = TerminalPanicked
			outcome.Err = &PanicError{Source: "core", Value: rec, Stack: debug.Stack()}
			err = outcome.Err
		} else {
			outcome.Status, outcome.ExternalState = classifyOutcome(ctx, execErr)
			outcome.Err = execErr
			outcome.Feedback = retryFeedback(execErr)
			err = execErr
		}

		timeout := r.SettleTimeout
		if timeout <= 0 {
			timeout = DefaultSettleTimeout
		}
		settleCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer cancel()

		if settleErr := r.Ledger.Settle(settleCtx, claim, outcome); settleErr != nil {
			wrapped := fmt.Errorf("sty: settle: %w", settleErr)
			if err != nil {
				err = errors.Join(err, wrapped)
			} else {
				err = wrapped
			}
			if r.OnSettleError != nil {
				r.OnSettleError(settleErr)
			}
		}
	}()

	execErr = r.Core.Execute(ctx, &RunState{Claim: claim})
	return execErr
}

// Run is the convenience form of Begin followed by RunClaim.
func (r *Runner) Run(ctx context.Context, key OperationKey, meta Metadata) error {
	claim, err := r.Begin(ctx, key, meta)
	if err != nil {
		return err
	}
	return r.RunClaim(ctx, claim)
}

func classifyOutcome(ctx context.Context, execErr error) (TerminalStatus, ExternalStateDisposition) {
	if execErr == nil {
		return TerminalAccepted, NothingLeft
	}
	if errors.Is(ctx.Err(), context.Canceled) && errors.Is(execErr, context.Canceled) {
		return TerminalCanceled, DispositionUnknown
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) && errors.Is(execErr, context.DeadlineExceeded) {
		return TerminalTimedOut, DispositionUnknown
	}
	return TerminalRejected, DispositionUnknown
}
