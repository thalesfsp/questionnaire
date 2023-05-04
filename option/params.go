// NOTE: Called `params` because package is already called `Option`.

package option

import (
	"github.com/thalesfsp/status"
)

//////
// Vars, consts, and types.
//////

// Func allows to set options.
type Func func(o *Options) error

// Answer is the answer to a question. It's a clone of `answer.Answer` otherwise
// circular dependency.
type Answer struct {
	// ID of the answer.
	ID string `json:",inline"`

	// QuestionID is the ID of the question at the time of the answer.
	QuestionID string `json:"questionId" validate:"required"`

	// QuestionLabel is the label of the question at the time of the answer.
	QuestionLabel string `json:"questionLabel" validate:"required"`

	// QuestionWeight is the weight of the option at the time of the answer.
	QuestionWeight int `json:"questionWeight"`

	// OptionID is the ID of the option at the time of the answer.
	OptionID string `json:"optionId"`

	// OptionLabel is the label of the option at the time of the answer.
	OptionLabel string `json:"optionLabel"`

	// OptionValue is the value at the time of the answer.
	OptionValue interface{} `json:"optionValue"`

	// OptionWeight is the weight of the option at the time of the answer.
	OptionWeight int `json:"optionWeight"`
}

// ForwardFunc runs in the `Forward` method.
type ForwardFunc func(a Answer) error

// NextQuestionFunc is a function which determines the next question.
type NextQuestionFunc func() string

// Options contains the fields shared between request's options.
type Options struct {
	// Label is the label of the option.
	Label string `json:"label"`

	// NextQuestion is the next question ID.
	NextQuestionID string `json:"nextQuestionID"`

	// state sets the state of the option.
	State status.Status `json:"state"`

	// Weight is the weight of the option.
	Weight int `json:"weight"`

	// Runs on `Forward`.
	forwardFunc ForwardFunc `json:"-"`
}

// WithForwardFunc sets `answer` the hook function.
func WithForwardFunc(f ForwardFunc) Func {
	return func(o *Options) error {
		o.forwardFunc = f

		return nil
	}
}

// WithLabel sets the label of the option.
func WithLabel(label string) Func {
	return func(o *Options) error {
		o.Label = label

		return nil
	}
}

// WithNextQuestionID sets the next question ID.
func WithNextQuestionID(id string) Func {
	return func(o *Options) error {
		o.NextQuestionID = id

		return nil
	}
}

// WithNextQuestionFunc sets the next question function.
func WithNextQuestionFunc(f NextQuestionFunc) Func {
	return func(o *Options) error {
		o.NextQuestionID = f()

		return nil
	}
}

// WithState sets the state of the option.
func WithState(s status.Status) Func {
	return func(o *Options) error {
		if s != status.None {
			o.State = s
		}

		return nil
	}
}

// WithWeight sets the weight of the option.
func WithWeight(w int) Func {
	return func(o *Options) error {
		o.Weight = w

		return nil
	}
}