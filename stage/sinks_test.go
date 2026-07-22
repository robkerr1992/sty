package stage

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestConsumeAllWithPolicy(t *testing.T) {
	ctx := context.Background()

	t.Run("fail fast runs in order and preserves identity", func(t *testing.T) {
		wantErr := errors.New("stop")
		var order []int
		sinks := []Sink[string]{
			SinkFunc[string](func(context.Context, string) error { order = append(order, 0); return nil }),
			SinkFunc[string](func(context.Context, string) error { order = append(order, 1); return wantErr }),
			SinkFunc[string](func(context.Context, string) error { order = append(order, 2); return nil }),
		}

		err := ConsumeAll(ctx, sinks, "value")
		if err != wantErr || !reflect.DeepEqual(order, []int{0, 1}) {
			t.Fatalf("ConsumeAll() error = %v, order = %v", err, order)
		}
	})

	t.Run("best effort runs all and joins failures", func(t *testing.T) {
		firstErr := errors.New("first")
		secondErr := errors.New("second")
		var order []int
		sinks := []Sink[string]{
			SinkFunc[string](func(context.Context, string) error { order = append(order, 0); return firstErr }),
			SinkFunc[string](func(context.Context, string) error { order = append(order, 1); return nil }),
			SinkFunc[string](func(context.Context, string) error { order = append(order, 2); return secondErr }),
		}

		err := ConsumeAllWithPolicy(ctx, BestEffort, sinks, "value")
		if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) || !reflect.DeepEqual(order, []int{0, 1, 2}) {
			t.Fatalf("ConsumeAllWithPolicy() error = %v, order = %v", err, order)
		}
	})

	t.Run("best effort wraps one failure", func(t *testing.T) {
		wantErr := errors.New("only")
		sinks := []Sink[string]{SinkFunc[string](func(context.Context, string) error { return wantErr })}

		err := ConsumeAllWithPolicy(ctx, BestEffort, sinks, "value")
		if err == wantErr || !errors.Is(err, wantErr) {
			t.Fatalf("ConsumeAllWithPolicy() error = %v", err)
		}
	})

	for _, tc := range []struct {
		name  string
		sinks []Sink[string]
	}{
		{name: "empty", sinks: []Sink[string]{}},
		{name: "nil", sinks: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ConsumeAllWithPolicy(ctx, BestEffort, tc.sinks, "value"); err != nil {
				t.Fatalf("ConsumeAllWithPolicy() error = %v", err)
			}
		})
	}

	t.Run("nil sink element panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("ConsumeAllWithPolicy() did not panic on nil sink")
			}
		}()
		_ = ConsumeAll(ctx, []Sink[string]{nil}, "value")
	})

	t.Run("unknown policy is fail fast", func(t *testing.T) {
		wantErr := errors.New("stop")
		calls := 0
		sinks := []Sink[string]{
			SinkFunc[string](func(context.Context, string) error { calls++; return wantErr }),
			SinkFunc[string](func(context.Context, string) error { calls++; return nil }),
		}

		err := ConsumeAllWithPolicy(ctx, SinkFailurePolicy(42), sinks, "value")
		if err != wantErr || calls != 1 {
			t.Fatalf("ConsumeAllWithPolicy() error = %v, calls = %d", err, calls)
		}
	})
}
