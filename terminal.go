package sty

import (
	"fmt"
	"time"
)

// TerminalStatus is the ledger-side terminal taxonomy. It is distinct from
// OutcomeStatus because lifecycle settlement also represents orphaning.
type TerminalStatus int

const (
	TerminalUnknown TerminalStatus = iota
	TerminalAccepted
	TerminalAcceptedWithCaveats
	TerminalRejected
	TerminalExhausted
	TerminalCanceled
	TerminalTimedOut
	TerminalPanicked
	// TerminalOrphaned is asserted only by Ledger.ReapStale, never by Runner.
	TerminalOrphaned
)

func (s TerminalStatus) String() string {
	switch s {
	case TerminalUnknown:
		return "TerminalUnknown"
	case TerminalAccepted:
		return "TerminalAccepted"
	case TerminalAcceptedWithCaveats:
		return "TerminalAcceptedWithCaveats"
	case TerminalRejected:
		return "TerminalRejected"
	case TerminalExhausted:
		return "TerminalExhausted"
	case TerminalCanceled:
		return "TerminalCanceled"
	case TerminalTimedOut:
		return "TerminalTimedOut"
	case TerminalPanicked:
		return "TerminalPanicked"
	case TerminalOrphaned:
		return "TerminalOrphaned"
	default:
		return fmt.Sprintf("TerminalStatus(%d)", s)
	}
}

// ExternalStateDisposition records what external state remains independently
// of why a run ended.
type ExternalStateDisposition int

const (
	DispositionUnknown ExternalStateDisposition = iota
	NothingLeft
	ArtifactsLeft
)

func (d ExternalStateDisposition) String() string {
	switch d {
	case DispositionUnknown:
		return "DispositionUnknown"
	case NothingLeft:
		return "NothingLeft"
	case ArtifactsLeft:
		return "ArtifactsLeft"
	default:
		return fmt.Sprintf("ExternalStateDisposition(%d)", d)
	}
}

// TerminalOutcome is the durable lifecycle result written by Ledger.Settle.
type TerminalOutcome struct {
	Status        TerminalStatus
	ExternalState ExternalStateDisposition
	Err           error
	EndedAt       time.Time
	// Feedback is consumer-supplied text intended for the next attempt,
	// independent of Status. Empty is the common case.
	Feedback string
}
