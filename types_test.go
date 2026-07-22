package sty

import (
	"errors"
	"testing"
)

func TestEnumStrings(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"VerdictNone", VerdictNone.String(), "VerdictNone"},
		{"VerdictRetry", VerdictRetry.String(), "VerdictRetry"},
		{"VerdictAccept", VerdictAccept.String(), "VerdictAccept"},
		{"VerdictAcceptWithIssues", VerdictAcceptWithIssues.String(), "VerdictAcceptWithIssues"},
		{"VerdictReject", VerdictReject.String(), "VerdictReject"},
		{"Verdict out of range", Verdict(7).String(), "Verdict(7)"},
		{"OpUnknown", OpUnknown.String(), "OpUnknown"},
		{"OpVerified", OpVerified.String(), "OpVerified"},
		{"OpStepError", OpStepError.String(), "OpStepError"},
		{"OpPermanentError", OpPermanentError.String(), "OpPermanentError"},
		{"OpPanic", OpPanic.String(), "OpPanic"},
		{"OpOutcome out of range", OpOutcome(7).String(), "OpOutcome(7)"},
		{"OutcomeUnknown", OutcomeUnknown.String(), "OutcomeUnknown"},
		{"OutcomeAccepted", OutcomeAccepted.String(), "OutcomeAccepted"},
		{"OutcomeAcceptedWithIssues", OutcomeAcceptedWithIssues.String(), "OutcomeAcceptedWithIssues"},
		{"OutcomeRejected", OutcomeRejected.String(), "OutcomeRejected"},
		{"OutcomeExhausted", OutcomeExhausted.String(), "OutcomeExhausted"},
		{"OutcomeCanceled", OutcomeCanceled.String(), "OutcomeCanceled"},
		{"OutcomeTimedOut", OutcomeTimedOut.String(), "OutcomeTimedOut"},
		{"OutcomePanicked", OutcomePanicked.String(), "OutcomePanicked"},
		{"OutcomeStatus out of range", OutcomeStatus(8).String(), "OutcomeStatus(8)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("String() = %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestEnumZeroValues(t *testing.T) {
	if Verdict(0) != VerdictNone {
		t.Fatal("Verdict zero value is not VerdictNone")
	}
	if OpOutcome(0) != OpUnknown {
		t.Fatal("OpOutcome zero value is not OpUnknown")
	}
	if OutcomeStatus(0) != OutcomeUnknown {
		t.Fatal("OutcomeStatus zero value is not OutcomeUnknown")
	}
}

func TestGenericInstantiations(t *testing.T) {
	tests := []struct {
		name    string
		attempt any
		outcome any
	}{
		{"strings", AttemptResult[string, string]{}, Outcome[string, string]{}},
		{"errors", AttemptResult[[]byte, error]{Issues: []error{errors.New("issue")}}, Outcome[[]byte, error]{}},
		{"structs", AttemptResult[struct{ Value int }, struct{ Code string }]{}, Outcome[struct{ Value int }, struct{ Code string }]{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.attempt == nil || tc.outcome == nil {
				t.Fatal("generic instantiation unexpectedly nil")
			}
		})
	}
}
