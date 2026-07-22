package sty

import (
	"fmt"
	"time"
)

// Verdict is the kernel's decision about an attempt's output. Its integer
// values are not a wire or storage contract; translators own persisted maps.
type Verdict int

const (
	VerdictNone Verdict = iota
	VerdictRetry
	VerdictAccept
	VerdictAcceptWithIssues
	VerdictReject
)

func (v Verdict) String() string {
	switch v {
	case VerdictNone:
		return "VerdictNone"
	case VerdictRetry:
		return "VerdictRetry"
	case VerdictAccept:
		return "VerdictAccept"
	case VerdictAcceptWithIssues:
		return "VerdictAcceptWithIssues"
	case VerdictReject:
		return "VerdictReject"
	default:
		return fmt.Sprintf("Verdict(%d)", v)
	}
}

// OpOutcome is the operational outcome of an attempt, independent of its
// output verdict. Its integer values are not a wire or storage contract.
type OpOutcome int

const (
	OpUnknown OpOutcome = iota
	OpVerified
	OpStepError
	OpPermanentError
	OpPanic
)

func (o OpOutcome) String() string {
	switch o {
	case OpUnknown:
		return "OpUnknown"
	case OpVerified:
		return "OpVerified"
	case OpStepError:
		return "OpStepError"
	case OpPermanentError:
		return "OpPermanentError"
	case OpPanic:
		return "OpPanic"
	default:
		return fmt.Sprintf("OpOutcome(%d)", o)
	}
}

// AttemptResult represents every attempt, including operational failures.
type AttemptResult[Out any, Iss any] struct {
	RunID   string
	Attempt int
	// StartedAt is UTC at the producer.
	StartedAt        time.Time
	Duration         time.Duration
	FeedbackInjected bool
	Output           Out
	Issues           []Iss
	Verdict          Verdict
	Op               OpOutcome
	Err              error
}
