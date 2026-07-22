package sty

import "time"

// Intent is a unit of work entering an agent execution.
type Intent struct {
	ID   string
	Text string
	// Meta is small and immutable by convention.
	Meta map[string]string
	// Created is UTC at the producer.
	Created time.Time
}

// PreparedPrompt is provisional. It sits in the settled interface chain, but
// its first real producer must validate or amend this shape.
type PreparedPrompt struct {
	IntentID string
	System   string
	User     string
	Inputs   []PromptInput
}

// PromptInput is a named value used to prepare a prompt.
type PromptInput struct {
	Name  string
	Value string
}
