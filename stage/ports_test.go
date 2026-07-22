package stage

import (
	"context"
	"errors"
	"testing"
)

func TestFunctionAdapters(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("adapter error")

	t.Run("source", func(t *testing.T) {
		calls := 0
		var source Source[string] = SourceFunc[string](func(gotCtx context.Context) (string, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("context was not passed through")
			}
			return "loaded", wantErr
		})

		got, err := source.Load(ctx)
		if calls != 1 || got != "loaded" || err != wantErr {
			t.Fatalf("Load() = (%q, %v), calls = %d", got, err, calls)
		}
	})

	t.Run("query", func(t *testing.T) {
		calls := 0
		var query QueryStage[int, string] = QueryFunc[int, string](func(gotCtx context.Context, in int) (string, error) {
			calls++
			if gotCtx != ctx {
				t.Fatal("context was not passed through")
			}
			if in != 42 {
				t.Fatalf("input = %d, want 42", in)
			}
			return "queried", wantErr
		})

		got, err := query.Query(ctx, 42)
		if calls != 1 || got != "queried" || err != wantErr {
			t.Fatalf("Query() = (%q, %v), calls = %d", got, err, calls)
		}
	})

	t.Run("sink", func(t *testing.T) {
		calls := 0
		var sink Sink[int] = SinkFunc[int](func(gotCtx context.Context, in int) error {
			calls++
			if gotCtx != ctx {
				t.Fatal("context was not passed through")
			}
			if in != 42 {
				t.Fatalf("input = %d, want 42", in)
			}
			return wantErr
		})

		err := sink.Consume(ctx, 42)
		if calls != 1 || err != wantErr {
			t.Fatalf("Consume() error = %v, calls = %d", err, calls)
		}
	})
}
