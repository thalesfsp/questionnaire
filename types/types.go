package types

//////
// Var, const, and types.
//////

// Type are the possible types of a question.
type Type string

const (
	Logical        Type = "logical"
	MultipleSelect Type = "multiple-select"
	SingleSelect   Type = "single-select"
	Text           Type = "text"
)

//////
// Methods.
//////

// String implements the Stringer interface.
func (t Type) String() string {
	return string(t)
}
