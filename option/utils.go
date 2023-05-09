package option

import (
	"github.com/thalesfsp/questionnaire/internal/shared"
)

// MapToOption converts a map[string]any to Option[T].
func MapToOption[T shared.N](m map[string]any) (Option[T], error) {
	// Marshall
	b, err := shared.Marshal(m)
	if err != nil {
		return Option[T]{}, err
	}

	// Unmarshall
	var o Option[T]
	if err := shared.Unmarshal(b, &o); err != nil {
		return Option[T]{}, err
	}

	return o, nil
}

// AnyToOption converts any to the proper Option[Type] then to any.
//
//nolint:gocritic,forcetypeassert,gosimple
func AnyToOption(v any) any {
	switch v.(type) {
	case Option[int]:
		return v.(Option[int])
	case Option[bool]:
		return v.(Option[bool])
	case Option[string]:
		return v.(Option[string])
	case Option[float32]:
		return v.(Option[float32])
	case Option[float64]:
		return v.(Option[float64])
	case Option[[]int]:
		return v.(Option[[]int])
	case Option[[]bool]:
		return v.(Option[[]bool])
	case Option[[]string]:
		return v.(Option[[]string])
	case Option[[]float32]:
		return v.(Option[[]float32])
	case Option[[]float64]:
		return v.(Option[[]float64])
	}

	return nil
}
