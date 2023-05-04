package answer

import (
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/question"
)

//////
// Var, const, and types.
//////

// IAnswer is an interface for answers.
type IAnswer interface {
	// GetID returns the ID of the answer.
	GetID() string

	// GetQuestion returns the question at the time of the answer.
	GetQuestion() question.Question

	// GetOption returns the option at the time of the answer.
	GetOption() option.IOption
}
