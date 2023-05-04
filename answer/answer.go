package answer

import (
	"github.com/thalesfsp/configurer/util"
	"github.com/thalesfsp/questionnaire/common"
	"github.com/thalesfsp/questionnaire/errorcatalog"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/question"
)

//////
// Var, const, and types.
//////

// Answer is an answer to a question.
type Answer[T any] struct {
	common.Common `json:",inline"`

	// Question at the time of the answer.
	Question question.Question `json:"question"`

	// Option is the Option at the time of the answer.
	Option option.IOption `json:"option"`
}

//////
// Implements the IAnswer interface.
//////

// GetID returns the ID of the option.
func (a Answer[T]) GetID() string {
	return a.Common.ID
}

// GetQuestion returns the question selected for the answer.
func (a Answer[T]) GetQuestion() question.Question {
	return a.Question
}

// GetOption returns the option at the time of the answer.
func (a Answer[T]) GetOption() option.IOption {
	return a.Option
}

// Validate answer.
func (a Answer[T]) Validate() error {
	if a.Question.Meta.Required && a.Option == nil {
		return errorcatalog.Catalog.MustGet(errorcatalog.ErrAnswerOptionRequired)
	}

	return nil
}

//////
// Factory.
//
// NOTE: All entities (Option, Question, etc) should have a factory which
// runs Process().
//////

// New creates a new Answer.
func New[T any](
	q question.Question,
	option option.IOption,
) (Answer[T], error) {
	a := Answer[T]{
		Common: common.Common{
			// It uses the question ID as the answer ID to help with the search.
			ID: q.GetID(),
		},

		Question: q,
		Option:   option,
	}

	if err := util.Process(&a); err != nil {
		return Answer[T]{}, err
	}

	return a, nil
}

// MustNew creates a new Answer and panics if there's an error.
func MustNew[T any](
	q question.Question,
	option option.IOption,
) Answer[T] {
	a, err := New[T](q, option)
	if err != nil {
		panic(err)
	}

	return a
}
