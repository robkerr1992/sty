package sty

import (
	"context"
	"testing"

	"github.com/robkerr1992/sty/stage"
)

type recordingAttemptSink struct {
	called bool
}

func (s *recordingAttemptSink) Consume(context.Context, AttemptResult[string, int]) error {
	s.called = true
	return nil
}

func TestAttemptSinkInterface(t *testing.T) {
	var sink AttemptSink[string, int] = &recordingAttemptSink{}
	if err := sink.Consume(context.Background(), AttemptResult[string, int]{}); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if !sink.(*recordingAttemptSink).called {
		t.Fatal("Consume() did not invoke sink")
	}
}

func TestPolicyAttemptSinkPairsSinkAndPolicy(t *testing.T) {
	sink := &recordingAttemptSink{}
	entry := PolicyAttemptSink[string, int]{Sink: sink, Policy: stage.BestEffort}

	if entry.Sink != sink {
		t.Fatal("Sink was not preserved")
	}
	if entry.Policy != stage.BestEffort {
		t.Fatalf("Policy = %v, want BestEffort", entry.Policy)
	}
}
