// NOTE: Called `params` for consistency (see option/params.go comment).

package question

import "github.com/thalesfsp/questionnaire/option"

//////
// Vars, consts, and types.
//////

// Func allows to set options.
type Func func(o *Meta) error

// NextQuestionFunc is a function which determines the next question.
type NextQuestionFunc func() string

// WithMeta set the whole meta. Replaces whatever was set before.
func WithMeta(meta *Meta) Func {
	return func(m *Meta) error {
		*m = *meta

		return nil
	}
}

// WithImageURL sets the question image URL.
func WithImageURL(imageURL string) Func {
	return func(m *Meta) error {
		m.ImageURL = imageURL

		return nil
	}
}

// WithRequired sets the question as required.
func WithRequired(required bool) Func {
	return func(m *Meta) error {
		m.Required = required

		return nil
	}
}

// WithWeight sets the question weight.
func WithWeight(weight int) Func {
	return func(m *Meta) error {
		m.Weight = weight

		return nil
	}
}

// WithOption add an option to the question.
func WithOption(opts ...option.IOption) Func {
	return func(m *Meta) error {
		if m.options == nil {
			m.options = []option.IOption{}
		}

		m.options = append(m.options, opts...)

		return nil
	}
}
