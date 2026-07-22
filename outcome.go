package sty

import (
	"fmt"
	"time"
)

// OutcomeStatus describes a run's self-reported terminal status. Its zero
// value is unknown so an uninitialized Outcome never reads as success. Its
// integer values are not a wire or storage contract. Orphaning is ledger-side
// state represented by Phase 3's distinct TerminalOutcome type.
type OutcomeStatus int

const (
	OutcomeUnknown OutcomeStatus = iota
	OutcomeAccepted
	OutcomeAcceptedWithIssues
	OutcomeRejected
	OutcomeExhausted
	OutcomeCanceled
	OutcomeTimedOut
	OutcomePanicked
)

func (s OutcomeStatus) String() string {
	switch s {
	case OutcomeUnknown:
		return "OutcomeUnknown"
	case OutcomeAccepted:
		return "OutcomeAccepted"
	case OutcomeAcceptedWithIssues:
		return "OutcomeAcceptedWithIssues"
	case OutcomeRejected:
		return "OutcomeRejected"
	case OutcomeExhausted:
		return "OutcomeExhausted"
	case OutcomeCanceled:
		return "OutcomeCanceled"
	case OutcomeTimedOut:
		return "OutcomeTimedOut"
	case OutcomePanicked:
		return "OutcomePanicked"
	default:
		return fmt.Sprintf("OutcomeStatus(%d)", s)
	}
}

// Outcome is the result of executing an intent.
type Outcome[Out any, Iss any] struct {
	IntentID string
	Status   OutcomeStatus
	Output   Out
	Issues   []Iss
	Attempts []AttemptResult[Out, Iss]
	// StartedAt and EndedAt are UTC at the producer.
	StartedAt time.Time
	EndedAt   time.Time
	Err       error
}
