package sty

import "fmt"

// PanicError mirrors pig.PanicError's Source/Value/Stack shape but is a
// distinct type because sty does not import fleet modules. It represents a
// panic recovered by Runner around Core.Execute.
type PanicError struct {
	Source string
	Value  any
	Stack  []byte
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("%s panic: %v", e.Source, e.Value)
}
