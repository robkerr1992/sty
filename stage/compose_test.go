package stage

import (
	"context"
	"errors"
	"testing"
)

func TestThen(t *testing.T) {
	ctx := context.Background()

	t.Run("first error short circuits", func(t *testing.T) {
		wantErr := errors.New("first")
		secondCalls := 0
		composed := Then(
			QueryFunc[int, string](func(context.Context, int) (string, error) { return "partial", wantErr }),
			QueryFunc[string, int](func(context.Context, string) (int, error) {
				secondCalls++
				return 99, nil
			}),
		)

		got, err := composed.Query(ctx, 1)
		if got != 0 || err != wantErr || secondCalls != 0 {
			t.Fatalf("Query() = (%d, %v), second calls = %d", got, err, secondCalls)
		}
	})

	t.Run("second error preserves identity", func(t *testing.T) {
		wantErr := errors.New("second")
		composed := Then(
			QueryFunc[int, string](func(context.Context, int) (string, error) { return "next", nil }),
			QueryFunc[string, int](func(context.Context, string) (int, error) { return 99, wantErr }),
		)

		got, err := composed.Query(ctx, 1)
		if got != 0 || err != wantErr {
			t.Fatalf("Query() = (%d, %v)", got, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		composed := Then(
			QueryFunc[int, string](func(context.Context, int) (string, error) { return "next", nil }),
			QueryFunc[string, int](func(context.Context, string) (int, error) { return 99, nil }),
		)

		got, err := composed.Query(ctx, 1)
		if got != 99 || err != nil {
			t.Fatalf("Query() = (%d, %v)", got, err)
		}
	})
}

func TestFromSource(t *testing.T) {
	ctx := context.Background()

	t.Run("load error skips query", func(t *testing.T) {
		wantErr := errors.New("load")
		queryCalls := 0
		source := FromSource(
			SourceFunc[int](func(context.Context) (int, error) { return 42, wantErr }),
			QueryFunc[int, string](func(context.Context, int) (string, error) {
				queryCalls++
				return "result", nil
			}),
		)

		got, err := source.Load(ctx)
		if got != "" || err != wantErr || queryCalls != 0 {
			t.Fatalf("Load() = (%q, %v), query calls = %d", got, err, queryCalls)
		}
	})

	t.Run("query error preserves identity", func(t *testing.T) {
		wantErr := errors.New("query")
		source := FromSource(
			SourceFunc[int](func(context.Context) (int, error) { return 42, nil }),
			QueryFunc[int, string](func(context.Context, int) (string, error) { return "partial", wantErr }),
		)

		got, err := source.Load(ctx)
		if got != "partial" || err != wantErr {
			t.Fatalf("Load() = (%q, %v)", got, err)
		}
	})
}
