package question

import (
	"github.com/thalesfsp/configurer/util"
	"github.com/thalesfsp/go-common-types/safeorderedmap"
	"github.com/thalesfsp/questionnaire/common"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/types"
)

//////
// Consts, vars, and types.
//////

// Meta enriches the question with metadata. Add here anything you need.
type Meta struct {
	// ID of the question.
	ID string `json:"id"`

	// ImageURL is the URL of the image.
	ImageURL string `json:"url"`

	// Index is the index of the question.
	Index int `json:"index"`

	// Required is a flag to indicate if the question is required.
	Required bool `json:"required" default:"false"`

	// Weight is the weight of the question.
	Weight int `json:"weight"`

	// Options is a list of options for the question to be answered.
	options []any `json:"-"`
}

// Question with options to be answered.
type Question struct {
	common.Common `json:",inline"`

	// Meta is the metadata of the question.
	Meta Meta `json:"meta"`

	// Label is the question.
	Label string `json:"label"`

	// Options is a list of options for the question to be answered.
	Options *safeorderedmap.SafeOrderedMap[any] `json:"options"`

	// PreviousQuestionID is the ID of the previous question.
	PreviousQuestionID string `json:"-"`

	// Type of the question.
	Type types.Type `json:"type"`
}

//////
// Methods.
//////

// GetID returns the ID of the question.
func (q *Question) GetID() string {
	return q.Common.ID
}

// GetIndex returns the index of the question.
func (q *Question) GetIndex() int {
	return q.Meta.Index
}

// SetIndex sets the index of the question.
func (q *Question) SetIndex(index int) {
	q.Meta.Index = index
}

// AddOption adds an option to the question.
func AddOption[T shared.N](q *Question, o ...option.Option[T]) {
	for _, opt := range o {
		q.Options.Add(opt.GetID(), opt)
	}
}

// GetOption returns an option from the question.
//
//nolint:forcetypeassert
func GetOption[T shared.N](q Question, id string) (option.Option[T], error) {
	opt, _ := q.Options.Get(id)

	return option.MapToOption[T](opt.(map[string]any))
}

//////
// Factory.
//
// NOTE: All entities (Option, Question, etc) should have a factory which
// runs Process().
//////

// New creates a new Question.
//
//nolint:forcetypeassert
func New[T shared.N](
	id string,
	label string,
	t types.Type,
	params ...Func,
) (Question, error) {
	m := Meta{
		ImageURL: "",
		Required: false,
		options:  []any{},
		Weight:   1,
	}

	// Applies params.
	for _, param := range params {
		if err := param(&m); err != nil {
			return Question{}, err
		}
	}

	//////
	// Add options to the question.
	//////

	// Initialize the options map.
	optsMap := safeorderedmap.New[any]()

	// Iterate over the m.options (any) and add to the map
	for _, o := range m.options {
		oTemp := o.(option.Option[T])

		optsMap.Add(oTemp.GetID(), oTemp)
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
func MustNew[T shared.N](
	id string,
	label string,
	t types.Type,
	params ...Func,
) Question {
	q, err := New[T](id, label, t, params...)
	if err != nil {
		panic(err)
	}

	return q
}
