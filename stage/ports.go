// Package stage provides domain-agnostic ports and composition helpers.
package stage

import "context"

// Source is a driving port where work enters. The pull port is generic:
// sty.Intent is one possible instantiation, while consumers with richer
// intents pull their own type rather than placing untyped data in metadata.
type Source[O any] interface {
	Load(ctx context.Context) (O, error)
}

// QueryStage is a CQS query: it returns data and performs no externally
// meaningful mutation. Effectful reads are permitted.
type QueryStage[I any, O any] interface {
	Query(ctx context.Context, in I) (O, error)
}

// Sink is a CQS command: it consumes a value and returns only an error.
type Sink[I any] interface {
	Consume(ctx context.Context, in I) error
}

// SourceFunc adapts a function to Source. A nil SourceFunc panics on
// invocation, matching the http.HandlerFunc convention for caller bugs.
type SourceFunc[O any] func(context.Context) (O, error)

// Load invokes f with ctx.
func (f SourceFunc[O]) Load(ctx context.Context) (O, error) { return f(ctx) }

// QueryFunc adapts a function to QueryStage. A nil QueryFunc panics on
// invocation, matching the http.HandlerFunc convention for caller bugs.
type QueryFunc[I any, O any] func(context.Context, I) (O, error)

// Query invokes f with ctx and in.
func (f QueryFunc[I, O]) Query(ctx context.Context, in I) (O, error) { return f(ctx, in) }

// SinkFunc adapts a function to Sink. A nil SinkFunc panics on invocation,
// matching the http.HandlerFunc convention for caller bugs.
type SinkFunc[I any] func(context.Context, I) error

// Consume invokes f with ctx and in.
func (f SinkFunc[I]) Consume(ctx context.Context, in I) error { return f(ctx, in) }
