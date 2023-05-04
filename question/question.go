package question

import (
	"github.com/thalesfsp/configurer/util"
	"github.com/thalesfsp/questionnaire/common"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/types"
)

//////
// Var, const, and types.
//////

// Meta enriches the question with metadata. Add here anything you need.
type Meta struct {
	// ImageURL is the URL of the image.
	ImageURL string `json:"url"`

	// Required is a flag to indicate if the question is required.
	Required bool `json:"required" default:"false"`

	// Weight is the weight of the question.
	Weight int `json:"weight"`

	// Options is a list of options for the question to be answered.
	options []option.IOption `json:"-"`
}

// Question with options to be answered.
type Question struct {
	common.Common `json:",inline"`

	// Meta is the metadata of the question.
	Meta Meta `json:"meta"`

	// Label is the question.
	Label string `json:"label" validate:"required"`

	// Options is a list of options for the question to be answered.
	Options *option.Map `json:"options"`

	// PreviousQuestionID is the ID of the previous question.
	PreviousQuestionID string `json:"-"`

	// Type of the question.
	Type types.Type `json:"type" validate:"required"`
}

//////
// Methods.
//////

// GetID returns the ID of the question.
func (q *Question) GetID() string {
	return q.Common.ID
}

// AddOption adds options to the question.
func (q *Question) AddOption(o ...option.IOption) *Question {
	for _, opt := range o {
		q.Options.Store(opt.GetID(), opt)
	}

	return q
}

//////
// Factory.
//
// NOTE: All entities (Option, Question, etc) should have a factory which
// runs Process().
//////

// New creates a new questionnaire.
func New(
	id string,
	label string,
	t types.Type,
	params ...Func,
) (Question, error) {
	m := Meta{
		ImageURL: "",
		Required: false,
		options:  []option.IOption{},
		Weight:   1,
	}

	// Applies params.
	for _, param := range params {
		if err := param(&m); err != nil {
			return Question{}, err
		}
	}

	optsMap := option.NewMap()

	for _, o := range m.options {
		optsMap.Store(o.GetID(), o)
	}

	q := Question{
		Common: common.Common{
			ID: id,
		},

		Meta: m,

		Label:   label,
		Options: optsMap,
		Type:    t,
	}

	if err := util.Process(&q); err != nil {
		return Question{}, err
	}

	return q, nil
}

// MustNew creates a new questionnaire and panics if there's an error.
func MustNew(
	id string,
	label string,
	t types.Type,
	params ...Func,
) Question {
	q, err := New(id, label, t, params...)
	if err != nil {
		panic(err)
	}

	return q
}
