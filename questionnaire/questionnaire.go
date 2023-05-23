package questionnaire

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/thalesfsp/configurer/util"
	"github.com/thalesfsp/go-common-types/safeorderedmap"
	"github.com/thalesfsp/questionnaire/common"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/questionnaire/question"
)

//////
// Consts, vars, and types.
//////

// Questionnaire is a set of questions.
type Questionnaire struct {
	common.Common `json:",inline" bson:",inline"`

	// Hash is a hash based on SHA-256. The goal is to avoid data tampering.
	Hash string `json:"hash" bson:"hash"`

	// Questions is a list of questions.
	Questions *safeorderedmap.SafeOrderedMap[question.Question] `json:"questions" bson:"questions" validate:"required,dive,required"`

	// Title of the questionnaire.
	Title string `json:"title" bson:"title" validate:"required"`
}

// generateHash generates a hash based on SHA-256. The goal is to avoid data
// tampering.
func (q *Questionnaire) generateHash() (string, error) {
	// Convert q to string.
	b, err := shared.Marshal(q)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	hashString := hex.EncodeToString(hash[:])

	return hashString, nil
}

//////
// Factory.
//////

// New creates a new questionnaire.
func New(title string, questions ...question.Question) (*Questionnaire, error) {
	questionMap := safeorderedmap.New[question.Question]()

	for i, q := range questions {
		// Set index.
		q.SetIndex(i)

		questionMap.Add(q.GetID(), q)
	}

	q := &Questionnaire{
		Title:     title,
		Questions: questionMap,
	}

	if err := util.Process(q); err != nil {
		return nil, err
	}

	h, err := q.generateHash()
	if err != nil {
		return nil, err
	}

	q.Hash = h

	return q, nil
}
