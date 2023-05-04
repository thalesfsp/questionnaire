package option

import (
	"github.com/thalesfsp/configurer/util"
	"github.com/thalesfsp/questionnaire/common"
	"github.com/thalesfsp/status"
)

//////
// Var, const, and types.
//////

// Option is a value for a question.
type Option[T any] struct {
	common.Common `json:",inline"`

	// Label is the label of the option.
	Label string `json:"label"`

	// Value is the value of the option.
	Value T `json:"value"`

	// Weight is the weight of the option.
	Weight int `json:"weight"`

	// NextQuestion is the next question ID.
	nextQuestionID string `json:"-"`

	// Runs on `Forward`.
	forwardFunc func(a Answer) error `json:"-"`

	// state sets the state of the option.
	state status.Status `json:"-"`
}

//////
// Implements the IOption interface.
//////

// GetID returns the ID of the option.
func (o Option[T]) GetID() string {
	return o.Common.ID
}

// GetAnswerFunc returns the answer function.
func (o Option[T]) GetAnswerFunc() ForwardFunc {
	return o.forwardFunc
}

// GetLabel returns the label of the option.
func (o Option[T]) GetLabel() string {
	return o.Label
}

// GetValue returns the value of the option.
func (o Option[T]) GetValue() interface{} {
	return o.Value
}

// GetState returns the state determiner function.
func (o Option[T]) GetState() status.Status {
	return o.state
}

// GetWeight returns the weight of the option.
func (o Option[T]) GetWeight() int {
	return o.Weight
}

// NextQuestionID returns next question index.
func (o Option[T]) NextQuestionID() string {
	return o.nextQuestionID
}

//////
// Factory.
//
// NOTE: All entities (Option, Question, etc) should have a factory which
// runs Process().
//////

// New creates a new Option.
func New[T any](value T, params ...Func) (Option[T], error) {
	p := &Options{
		State:  status.None,
		Weight: 1,
	}

	// Applies params.
	for _, param := range params {
		if err := param(p); err != nil {
			return Option[T]{}, err
		}
	}

	o := Option[T]{
		Label:          p.Label,
		Value:          value,
		Weight:         p.Weight,
		nextQuestionID: p.NextQuestionID,
		forwardFunc:    p.forwardFunc,
		state:          p.State,
	}

	if err := util.Process(&o); err != nil {
		return Option[T]{}, err
	}

	return o, nil
}

// MustNew creates a new Option and panics if there's an error.
func MustNew[T any](value T, p ...Func) Option[T] {
	o, err := New(value, p...)
	if err != nil {
		panic(err)
	}

	return o
}
