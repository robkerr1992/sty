package sty

import (
	"context"
	"errors"
	"time"
)

// Worker polls a PullableLedger through Runner.BeginNext and runs each claim
// through Runner.RunClaim. It imports no execution-engine packages.
type Worker struct {
	Runner     *Runner
	Meta       Metadata
	IdlePoll   time.Duration
	OnRunError func(error)
}

const DefaultIdlePoll = 2 * time.Second

// Run polls until ctx ends. Empty-queue and error paths wait IdlePoll; a
// successful claim and run polls again immediately.
func (w *Worker) Run(ctx context.Context) error {
	idle := w.IdlePoll
	if idle <= 0 {
		idle = DefaultIdlePoll
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		claim, err := w.Runner.BeginNext(ctx, w.Meta)
		if err != nil {
			if !errors.Is(err, ErrNoPendingWork) && w.OnRunError != nil {
				w.OnRunError(err)
			}
			if waitErr := w.sleep(ctx, idle); waitErr != nil {
				return waitErr
			}
			continue
		}
		if runErr := w.Runner.RunClaim(ctx, claim); runErr != nil && w.OnRunError != nil {
			w.OnRunError(runErr)
		}
	}
}

func (w *Worker) sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
