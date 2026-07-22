package sty

import "fmt"

// Phase separates query-only PRE stages from side-effecting POST stages.
type Phase int

const (
	PhasePre Phase = iota
	PhasePost
)

// PlanStage carries validation metadata. Plan is validation-only in Phase 3;
// actual stage execution remains inside each consumer's Core.Execute.
type PlanStage struct {
	Name                    string
	Phase                   Phase
	Order                   int
	Required                bool
	ExternalMutation        bool
	ShortCircuitsSettlement bool
}

// Plan declares ordered stages and names that must appear exactly once.
type Plan struct {
	Stages   []PlanStage
	Required []string
}

// Validate performs the four phase-plan checks in their specified order.
func (p Plan) Validate() error {
	seen := map[string]int{}
	for _, stage := range p.Stages {
		seen[stage.Name]++
	}
	for _, name := range p.Required {
		switch seen[name] {
		case 0:
			return fmt.Errorf("sty: phase plan: missing required stage %q", name)
		case 1:
		default:
			return fmt.Errorf("sty: phase plan: stage %q declared %d times, required exactly once", name, seen[name])
		}
	}

	type phaseOrder struct {
		phase Phase
		order int
	}
	byPhaseOrder := map[phaseOrder]string{}
	for _, stage := range p.Stages {
		key := phaseOrder{stage.Phase, stage.Order}
		if existing, ok := byPhaseOrder[key]; ok {
			return fmt.Errorf("sty: phase plan: duplicate phase/order pair (phase=%v order=%d): %q and %q", stage.Phase, stage.Order, existing, stage.Name)
		}
		byPhaseOrder[key] = stage.Name
	}

	for _, stage := range p.Stages {
		if stage.Phase == PhasePre && stage.ExternalMutation {
			return fmt.Errorf("sty: phase plan: PRE stage %q declares ExternalMutation", stage.Name)
		}
	}

	// This check is intentionally unconditional across all stages. It must run
	// before the POST-only ordering scan so PRE short-circuit declarations are
	// not silently skipped.
	for _, stage := range p.Stages {
		if stage.Phase != PhasePost && stage.ShortCircuitsSettlement {
			return fmt.Errorf("sty: phase plan: short-circuit stage %q is not in POST", stage.Name)
		}
	}

	var lastPostOrder int
	var lastPostName string
	var shortCircuit PlanStage
	haveShortCircuit := false
	for _, stage := range p.Stages {
		if stage.Phase != PhasePost {
			continue
		}
		if stage.Order >= lastPostOrder || lastPostName == "" {
			lastPostOrder = stage.Order
			lastPostName = stage.Name
		}
		if stage.ShortCircuitsSettlement {
			if haveShortCircuit {
				return fmt.Errorf("sty: phase plan: multiple short-circuit stages declared (%q, %q) — at most one is allowed", shortCircuit.Name, stage.Name)
			}
			haveShortCircuit = true
			shortCircuit = stage
		}
	}
	if haveShortCircuit && shortCircuit.Name != lastPostName {
		return fmt.Errorf("sty: phase plan: short-circuit stage %q is not last in POST (last is %q)", shortCircuit.Name, lastPostName)
	}
	return nil
}
