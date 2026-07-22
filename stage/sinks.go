package stage

import (
	"context"
	"errors"
)

// SinkFailurePolicy controls how a sink runner handles failures.
type SinkFailurePolicy int

const (
	// FailFast is deliberately the zero value: fail loud by default. It stops
	// at the first sink error and preserves that error's identity.
	FailFast SinkFailurePolicy = iota
	// BestEffort runs every sink and accumulates failures with errors.Join.
	BestEffort
)

// ConsumeAll is sugar for ConsumeAllWithPolicy(ctx, FailFast, sinks, in).
func ConsumeAll[I any](ctx context.Context, sinks []Sink[I], in I) error {
	return ConsumeAllWithPolicy(ctx, FailFast, sinks, in)
}

// ConsumeAllWithPolicy invokes sinks sequentially in slice order on the
// calling goroutine. Runners are not goroutine-safe; sinks own their thread
// safety. Empty or nil slices return nil, while a nil sink element panics as a
// caller bug. Unknown policies use FailFast. FailFast returns the original
// error unchanged. BestEffort uses errors.Join, which wraps even one error, so
// callers match joined failures with errors.Is rather than ==.
func ConsumeAllWithPolicy[I any](ctx context.Context, policy SinkFailurePolicy, sinks []Sink[I], in I) error {
	switch policy {
	case BestEffort:
		var failures []error
		for _, sink := range sinks {
			if err := sink.Consume(ctx, in); err != nil {
				failures = append(failures, err)
			}
		}
		return errors.Join(failures...)
	case FailFast:
		fallthrough
	default:
		for _, sink := range sinks {
			if err := sink.Consume(ctx, in); err != nil {
				return err
			}
		}
		return nil
	}
}
