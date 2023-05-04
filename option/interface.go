package option

import (
	"github.com/thalesfsp/status"
)

//////
// Var, const, and types.
//////

// IOption is an interface for options.
type IOption interface {
	// GetID returns the ID of the option.
	GetID() string

	// GetAnswerFunc returns the answer function.
	GetAnswerFunc() ForwardFunc

	// GetLabel returns the label of the option.
	GetLabel() string

	// GetValue returns the value of the option.
	GetValue() interface{}

	// GetState returns the state determiner function.
	GetState() status.Status

	// GetWeight returns the weight of the option.
	GetWeight() int

	// NextQuestionID returns next question index.
	NextQuestionID() string
}
