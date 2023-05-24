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
	OptionValue any `json:"optionValue"`

	// OptionWeight is the weight of the option at the time of the answer.
	OptionWeight int `json:"optionWeight"`
}

// NextQuestionFunc is a function which determines the next question.
type NextQuestionFunc func() string

// Options contains the fields shared between request's options.
type Options struct {
	// ID of the option.
	ID string `json:"id"`

	// Group of the option.
	Group string `json:"group"`

	// ImageURL is the URL of the image.
	ImageURL string `json:"url"`

	// Label is the label of the option.
	Label string `json:"label"`

	// NextQuestion is the next question ID.
	NextQuestionID string `json:"nextQuestionID"`

	// QuestionID is the ID of the question.
	QuestionID string `json:"questionID" bson:"questionID"`

	// state sets the state of the option.
	State status.Status `json:"state"`

	// Weight is the weight of the option.
	Weight int `json:"weight"`
}

// WithGroup sets the group of an option.
func WithGroup(group string) Func {
	return func(o *Options) error {
		o.Group = group

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

// WithID sets the ID of the option.
func WithID(id string) Func {
	return func(o *Options) error {
		o.ID = id

		return nil
	}
}

// WithQuestionID sets the weight of the option.
func WithQuestionID(id string) Func {
	return func(o *Options) error {
		o.QuestionID = id

		return nil
	}
}
