package stage

import "context"

// Then chains two query stages. If either stage fails, Then returns the zero
// value of C and the original error unwrapped. The second stage is not invoked
// when the first stage fails.
func Then[A, B, C any](ab QueryStage[A, B], bc QueryStage[B, C]) QueryStage[A, C] {
	return QueryFunc[A, C](func(ctx context.Context, in A) (C, error) {
		b, err := ab.Query(ctx, in)
		if err != nil {
			var zero C
			return zero, err
		}

		c, err := bc.Query(ctx, b)
		if err != nil {
			var zero C
			return zero, err
		}
		return c, nil
	})
}

// FromSource lifts a query stage over a source. The query is not invoked when
// loading fails, and errors from either component are returned unwrapped.
func FromSource[A, B any](src Source[A], q QueryStage[A, B]) Source[B] {
	return SourceFunc[B](func(ctx context.Context) (B, error) {
		a, err := src.Load(ctx)
		if err != nil {
			var zero B
			return zero, err
		}
		return q.Query(ctx, a)
	})
}
