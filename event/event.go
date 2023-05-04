package event

import (
	"github.com/thalesfsp/questionnaire/answer"
	"github.com/thalesfsp/questionnaire/question"
	"github.com/thalesfsp/questionnaire/questionnaire"
	"github.com/thalesfsp/status"
)

//////
// Const, var, and types.
//////

// Callback is a function that is called every time the state of the
// questionnaire changes.
type Callback func(Event)

// Event emitted every time the state of the questionnaire changes.
type Event struct {
	//////
	// Metadata.
	//////

	// PreviousQuestion is the previous question.
	PreviousQuestion question.Question `json:"previousQuestion"`

	// CurrentQuestion is the current question.
	CurrentQuestion question.Question `json:"currentQuestion"`

	// State is the current state of the questionnaire.
	State status.Status `json:"state"`

	// TotalAnswers is the total number of answers.
	TotalAnswers int `json:"totalAnswers"`

	// TotalQuestions is the total number of questions.
	TotalQuestions int `json:"totalQuestions"`

	//////
	// Data.
	//////

	// Answers is the list of answers.
	Answers *answer.Map `json:"answers"`

	// Questionnaire is the questionnaire.
	Questionnaire questionnaire.Questionnaire `json:"questionnaire"`

	// UserID is the ID of the user.
	UserID string `json:"userID"`
}
