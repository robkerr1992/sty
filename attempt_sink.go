package sty

import (
	"context"

	"github.com/robkerr1992/sty/stage"
)

// AttemptSink observes every attempt a kernel runs, verified or not. It is the
// promoted replacement for ad hoc telemetry and verified-only attempt hooks.
type AttemptSink[Out any, Iss any] interface {
	Consume(ctx context.Context, r AttemptResult[Out, Iss]) error
}

// PolicyAttemptSink pairs a sink with its failure-propagation policy. Slice
// order is registration order and must be preserved by consumers.
type PolicyAttemptSink[Out any, Iss any] struct {
	Sink   AttemptSink[Out, Iss]
	Policy stage.SinkFailurePolicy
}
