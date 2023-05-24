package option

import (
	"github.com/thalesfsp/configurer/util"
	"github.com/thalesfsp/questionnaire/common"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/status"
)

//////
// Consts, vars, and types.
//////

// Option is a value for a question.
type Option[T shared.N] struct {
	common.Common `json:",inline" bson:",inline"`

	// Label is the label of the option.
	Label string `json:"label" bson:"label"`

	// QuestionID is the ID of the question.
	QuestionID string `json:"questionID" bson:"questionID"`

	// State sets the State of the option.
	State status.Status `json:"state" bson:"state"`

	// Value is the value of the option.
	Value T `json:"value" bson:"value"`

	// Weight is the weight of the option.
	Weight int `json:"weight" bson:"weight"`

	// NextQuestion is the next question ID.
	nextQuestionID string `json:"-" bson:"-"`
}

//////
// Implements the IOption interface.
//////

// GetID returns the ID of the option.
func (o Option[T]) GetID() string {
	return o.Common.ID
}

// GetLabel returns the label of the option.
func (o Option[T]) GetLabel() string {
	return o.Label
}

// GetQuestionID returns the question ID of the option.
func (o Option[T]) GetQuestionID() string {
	return o.QuestionID
}

// SetQuestionID returns the question ID of the option.
func (o Option[T]) SetQuestionID(id string) Option[T] {
	o.QuestionID = id

	return o
}

// GetValue returns the value of the option.
func (o Option[T]) GetValue() T {
	return o.Value
}

// GetState returns the state determiner function.
func (o Option[T]) GetState() status.Status {
	return o.State
}

// GetWeight returns the weight of the option.
func (o Option[T]) GetWeight() int {
	return o.Weight
}

// NextQuestionID returns next question index.
func (o Option[T]) NextQuestionID() string {
	return o.nextQuestionID
}

// Get returns the value of the option.
func (o Option[T]) Get() interface{} {
	return o
}

//////
// Factory.
//
// NOTE: All entities (Option, Question, etc) should have a factory which
// runs Process().
//////

// New creates a new Option.
func New[T shared.N](value T, params ...Func) (Option[T], error) {
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
		Label:      p.Label,
		QuestionID: p.QuestionID,
		Value:      value,
		Weight:     p.Weight,

		nextQuestionID: p.NextQuestionID,
		State:          p.State,
	}

	// Sets the ID if any.
	if p.ID != "" {
		o.ID = p.ID
	}

	if err := util.Process(&o); err != nil {
		return Option[T]{}, err
	}

	return o, nil
}

// MustNew creates a new Option and panics if there's an error.
func MustNew[T shared.N](value T, p ...Func) Option[T] {
	o, err := New(value, p...)
	if err != nil {
		panic(err)
	}

	return o
}
