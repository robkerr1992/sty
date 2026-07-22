package sty

import (
	"context"
	"strings"
	"testing"
	"time"
)

type countingLedger struct{ begins int }

func (l *countingLedger) Begin(_ context.Context, key OperationKey, _ Metadata) (Claim, error) {
	l.begins++
	return Claim{Key: key, Attempt: 1}, nil
}
func (*countingLedger) Settle(context.Context, Claim, TerminalOutcome) error    { return nil }
func (*countingLedger) ReapStale(context.Context, time.Duration) (int64, error) { return 0, nil }

func TestPlanValidateFailuresPreventBegin(t *testing.T) {
	tests := []struct {
		name     string
		plan     Plan
		contains string
	}{
		{"missing required", Plan{Required: []string{"load"}}, "missing required stage \"load\""},
		{"required twice", Plan{Required: []string{"load"}, Stages: []PlanStage{{Name: "load", Phase: PhasePre, Order: 0}, {Name: "load", Phase: PhasePre, Order: 1}}}, "declared 2 times"},
		{"duplicate phase order", Plan{Stages: []PlanStage{{Name: "persistResult", Phase: PhasePost, Order: 0}, {Name: "publishCorpus", Phase: PhasePost, Order: 0}}}, "duplicate phase/order pair"},
		{"pre external mutation", Plan{Stages: []PlanStage{{Name: "query", Phase: PhasePre, ExternalMutation: true}}}, "PRE stage \"query\" declares ExternalMutation"},
		{"pre short circuit", Plan{Stages: []PlanStage{{Name: "query", Phase: PhasePre, ShortCircuitsSettlement: true}}}, "short-circuit stage \"query\" is not in POST"},
		{"short circuit not last", Plan{Stages: []PlanStage{{Name: "publish", Phase: PhasePost, Order: 0, ShortCircuitsSettlement: true}, {Name: "audit", Phase: PhasePost, Order: 1}}}, "not last in POST"},
		{"multiple short circuits", Plan{Stages: []PlanStage{{Name: "publish", Phase: PhasePost, Order: 0, ShortCircuitsSettlement: true}, {Name: "audit", Phase: PhasePost, Order: 1, ShortCircuitsSettlement: true}}}, "multiple short-circuit stages"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ledger := &countingLedger{}
			runner := Runner{Ledger: ledger, Core: coreFunc(func(context.Context, *RunState) error { return nil }), Plan: tc.plan}
			err := runner.Run(context.Background(), "key", nil)
			if err == nil || !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("Run() error = %v, want containing %q", err, tc.contains)
			}
			if ledger.begins != 0 {
				t.Fatalf("Ledger.Begin calls = %d, want 0", ledger.begins)
			}
		})
	}
}

func TestPlanValidateValidWithoutShortCircuit(t *testing.T) {
	plan := Plan{
		Required: []string{"load", "persist"},
		Stages: []PlanStage{
			{Name: "load", Phase: PhasePre, Order: 0, Required: true},
			{Name: "inspect", Phase: PhasePre, Order: 1},
			{Name: "persist", Phase: PhasePost, Order: 0, Required: true, ExternalMutation: true},
			{Name: "publish", Phase: PhasePost, Order: 1, ExternalMutation: true},
		},
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
