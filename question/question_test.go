package question

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/types"
)

func TestQuestion_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		want    []byte
		wantErr bool
	}{
		{
			name: "Should work",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oID := "oID1"

			q := MustNew[int]("id1", "Qlabel1", types.SingleSelect,
				WithOption(option.MustNew(1,
					option.WithGroup("g1"),
					option.WithLabel("Olabel1"),
					option.WithID(oID),
				)),
			)

			got, err := shared.Marshal(&q)
			assert.NoError(t, err)

			var q2 Question
			if err := shared.Unmarshal(got, &q2); err != nil {
				assert.NoError(t, err)
			}

			opt, err := GetOption[int](q2, oID)
			assert.NoError(t, err)
			assert.EqualValues(t, 1, opt.Value)

			assert.EqualValues(t, q.Meta.ImageURL, q2.Meta.ImageURL)
			assert.EqualValues(t, q.Meta.Required, q2.Meta.Required)
			assert.EqualValues(t, q.Meta.Weight, q2.Meta.Weight)
			assert.EqualValues(t, q.Label, q2.Label)
			assert.EqualValues(t, q.PreviousQuestionID, q2.PreviousQuestionID)
			assert.EqualValues(t, q.Type, q2.Type)
		})
	}
}
