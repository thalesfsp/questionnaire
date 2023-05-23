package fsm

import (
	"context"
	"expvar"
	"fmt"

	"github.com/thalesfsp/go-common-types/safeorderedmap"
	"github.com/thalesfsp/questionnaire/answer"
	"github.com/thalesfsp/questionnaire/errorcatalog"
	"github.com/thalesfsp/questionnaire/event"
	"github.com/thalesfsp/questionnaire/internal/customapm"
	"github.com/thalesfsp/questionnaire/internal/logging"
	"github.com/thalesfsp/questionnaire/internal/metrics"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/question"
	"github.com/thalesfsp/questionnaire/questionnaire"
	"github.com/thalesfsp/status"
	"github.com/thalesfsp/sypl"
	"github.com/thalesfsp/sypl/level"
	"github.com/thalesfsp/validation"
)

//////
// Consts, vars, and types.
//////

// Type is the type of the entity regarding the framework. It is used to for
// example, to identify the entity in the logs, metrics, and for tracing.
const (
	DefaultMetricCounterLabel = "counter"
	Name                      = "fsm"
	Type                      = "FiniteStateMachine"
)

// Callback is a function that is called every time the state of the
// questionnaire changes.
type Callback func(e event.Event, journal []event.Event)

// FiniteStateMachine is the Finite State Machine for the Questionnaire.
type FiniteStateMachine struct {
	// CurrentQuestion is the current question.
	CurrentQuestion question.Question `json:"currentQuestion" bson:"currentQuestion"`

	// CurrentQuestionID is the ID of the current question.
	CurrentQuestionID string `json:"currentQuestionID" bson:"currentQuestionID"`

	// CurrentQuestionIndex is the current question index.
	CurrentQuestionIndex int `json:"currentQuestionIndex" bson:"currentQuestionIndex"`

	// PreviousQuestionID is the ID of the previous question.
	PreviousQuestionID string `json:"previousQuestionID" bson:"previousQuestionID"`

	// TotalAnswers is the total number of answers.
	TotalAnswers int `json:"totalAnswers" bson:"totalAnswers"`

	// TotalQuestions is the total number of questions.
	TotalQuestions int `json:"totalQuestions" bson:"totalQuestions"`

	// Journal is the list of events.
	Journal []event.Event `json:"journal" bson:"journal"`

	// Logger.
	Logger sypl.ISypl `json:"-" bson:"-" validate:"required"`

	// Answers is the list of answers.
	Answers *safeorderedmap.SafeOrderedMap[answer.Answer] `json:"answers" bson:"answers"`

	// Questionnaire is the questionnaire.
	Questionnaire questionnaire.Questionnaire `json:"Questionnaire" bson:"Questionnaire" validate:"required"`

	// PreviousQuestion is the previous question.
	PreviousQuestion question.Question `json:"previousQuestion" bson:"previousQuestion"`

	// CurrentAnswer is the current answer.
	CurrentAnswer answer.Answer `json:"currentAnswer" bson:"currentAnswer"`

	// State is the current state of the questionnaire.
	State status.Status `json:"state" validate:"required" bson:"state"`

	// UserID is the ID of the user.
	UserID string `json:"userID" validate:"required" bson:"userID"`

	// Callback is the function that is called every time the state of the
	// questionnaire changes. For example: save the state to the database.
	callback Callback `json:"-" bson:"-"`

	// Metrics.
	counterBackward            *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterCompleted           *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterDone                *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterEmitted             *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterForward             *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterForwardFailed       *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterInitialized         *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterInstantiationFailed *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
	counterJump                *expvar.Int `json:"-" bson:"-" validate:"required,gte=0"`
}

//////
// Methods.
//////

// GetLogger returns the logger.
func (fsm *FiniteStateMachine) GetLogger() sypl.ISypl {
	return fsm.Logger
}

// AddToJournal adds an entry to the journal.
func (fsm *FiniteStateMachine) AddToJournal(e event.Event) {
	fsm.Journal = append(fsm.Journal, e)
}

// GetJournal returns the journal.
func (fsm *FiniteStateMachine) GetJournal() []event.Event {
	return fsm.Journal
}

// GetState returns the current state of the questionnaire.
func (fsm *FiniteStateMachine) GetState() status.Status {
	return fsm.State
}

//////
// Helpers.
//////

// Navigate goes back to the previous question, or jump to the specified one.
func (fsm *FiniteStateMachine) navigate(id string) *FiniteStateMachine {
	// Should do nothing if there's no previous question or if the FSM is in the
	// Started status.
	// if fsm.PreviousQuestionID == "" || fsm.State == status.Initialized {
	// 	return fsm
	// }

	// Load the current question - just to store reference.
	currentQst, _ := fsm.Questionnaire.Questions.Get(fsm.CurrentQuestionID)

	// Load the answer based on the previous question ID. This is what matters.
	finalID := fsm.PreviousQuestionID

	// Allows to jump to a specific question.
	if id != "" {
		finalID = id
	}

	aswr, ok := fsm.Answers.Get(finalID)
	if !ok { // Should do nothing if there's no answer.
		return fsm
	}

	// Load the answered question.
	answeredQst := aswr.GetQuestion()

	// Set the current question ID to the loaded question.
	fsm.CurrentQuestionID = answeredQst.GetID()

	// Set the current question.
	fsm.CurrentQuestion = answeredQst

	// Set the current question index.
	fsm.CurrentQuestionIndex = fsm.CurrentQuestion.GetIndex()

	// Set the previous question ID to the loaded question's previous question.
	fsm.PreviousQuestionID = answeredQst.PreviousQuestionID

	// Load the previous question.
	previousQst, _ := fsm.Questionnaire.Questions.Get(fsm.PreviousQuestionID)

	// Set the previous question.
	fsm.PreviousQuestion = previousQst

	// Set the current answer.
	finalAnswer, _ := fsm.Answers.Get(answeredQst.GetID())
	fsm.CurrentAnswer = finalAnswer

	// Ensure's the proper state is set.
	fsm.State = status.Runnning

	// Emit the state of the machine.
	fsm.Emit(currentQst, answeredQst)

	// Metrics: increment the counter.
	if id == "" {
		fsm.counterBackward.Add(1)
	} else {
		fsm.counterJump.Add(1)
	}

	return fsm
}

//////
// Special state machine methods.
//////

// Start the state machine.
func (fsm *FiniteStateMachine) Start() *FiniteStateMachine {
	// Transition to the answering status.
	fsm.State = status.Runnning

	// Load the first question.
	_, qst, _ := fsm.Questionnaire.Questions.First()

	// Set the current question ID.
	fsm.CurrentQuestionID = qst.GetID()

	// Set the current question.
	fsm.CurrentQuestion = qst

	// Set the current question index.
	fsm.CurrentQuestionIndex = fsm.CurrentQuestion.GetIndex()

	// Emit the state of the machine.
	fsm.Emit(question.Question{}, qst)

	// Observability: metrics.
	fsm.counterInitialized.Add(1)

	return fsm
}

// Backward goes back to the previous question.
func (fsm *FiniteStateMachine) Backward() *FiniteStateMachine {
	return fsm.navigate("")
}

// Jump to the specified question.
func (fsm *FiniteStateMachine) Jump(id string) *FiniteStateMachine {
	return fsm.navigate(id)
}

// Done the FSM setting the state to `Done`.
func (fsm *FiniteStateMachine) Done() *FiniteStateMachine {
	// Ensure's the proper state is set.
	fsm.State = status.Done

	// Emit the state of the machine.
	fsm.Dump()

	// Observability: metrics.
	fsm.counterDone.Add(1)

	return fsm
}

// Emit emits the state of the machine, also returning it. Do whatever you want
// with it :) (e.g. persist it). It includes the a dump of the journal, so you
// can use it to restore the state of the machine.
func (fsm *FiniteStateMachine) Emit(prevQst, currentQst question.Question) event.Event {
	aswr, _ := fsm.Answers.Get(currentQst.GetID())

	cQI, _, _ := fsm.Questionnaire.Questions.Index(currentQst.GetID())

	e := event.Event{
		CurrentQuestion:      currentQst,
		CurrentQuestionIndex: cQI,
		CurrentAnswer:        aswr,
		State:                fsm.State,
		TotalAnswers:         fsm.Answers.Size(),
		TotalQuestions:       fsm.Questionnaire.Questions.Size(),
		UserID:               fsm.UserID,
		PreviousQuestion:     prevQst,
		Answers:              fsm.Answers,
		Questionnaire:        fsm.Questionnaire,
	}

	// Emit the state of the machine.
	if fsm.callback != nil {
		fsm.callback(e, fsm.GetJournal())
	}

	// Add the entry to the journal.
	fsm.AddToJournal(e)

	// Observability: metrics.
	fsm.counterEmitted.Add(1)

	// Observability: log.
	fsm.GetLogger().Debuglnf("%+v", e)

	return e
}

// Dump returns the current state of the machine.
func (fsm *FiniteStateMachine) Dump() event.Event {
	// Make sure to emit the latest state
	previousQuestion, _ := fsm.Questionnaire.Questions.Get(fsm.PreviousQuestionID)
	currentQuestion, _ := fsm.Questionnaire.Questions.Get(fsm.CurrentQuestionID)

	return fsm.Emit(previousQuestion, currentQuestion)
}

// SetCallback sets the callback to be called when the state of the machine
// changes.
func (fsm *FiniteStateMachine) SetCallback(cb Callback) {
	fsm.callback = cb
}

// Forward the current question with the given option.
//
//nolint:nestif
func Forward[T shared.N](ctx context.Context, fsm *FiniteStateMachine, opt option.Option[T]) error {
	// Ensure's the proper state is set.
	fsm.State = status.Runnning

	//////
	// Determine the current question.
	//////

	// Retrieves the current question from the Questionnaire.
	qst, _ := fsm.Questionnaire.Questions.Get(fsm.CurrentQuestionID)

	// Sets the previous question ID to the current question ID.
	fsm.PreviousQuestionID = qst.GetID()

	//////
	// Deal with answer.
	//
	// NOTE: The following are all the possible types of answers. Update
	// accordingly.
	//////

	// Create the answer.
	aswr, err := answer.New(qst, opt)
	if err != nil {
		return err
	}

	// Add answer to the list.
	fsm.Answers.Add(aswr.GetID(), aswr)

	// If all questions have been answered, transition to the completed status.
	if fsm.Answers.Size() == fsm.Questionnaire.Questions.Size() {
		fsm.State = status.Completed
	}

	//////
	// Deal with the next question.
	//////

	// Get the next question ID.
	nextQstID := opt.NextQuestionID()

	if nextQstID != "" {
		// Loads the question from the Questionnaire.
		nextQst, _ := fsm.Questionnaire.Questions.Get(nextQstID)

		//////
		// Deal with setting the previous question.
		//////

		// Sets the previous question ID.
		if nextQst.PreviousQuestionID == "" {
			nextQst.PreviousQuestionID = qst.GetID()
		}

		// Update questionnaire's questions with the updated one persisting the
		// change.
		fsm.Questionnaire.Questions.Add(nextQst.GetID(), nextQst)

		//////
		// Deal with settings the current question ID, and determining the status.
		//////

		// Update the current question ID.
		fsm.CurrentQuestionID = nextQst.GetID()

		// Set the current question.
		fsm.CurrentQuestion = nextQst

		// Set the current question index.
		fsm.CurrentQuestionIndex = fsm.CurrentQuestion.GetIndex()

		// Optionally, set the state based on the option (answer).
		if opt.GetState() != status.None {
			fsm.State = opt.GetState()
		}

		// Emit the state of the machine.
		fsm.Emit(qst, nextQst)
	} else {
		if opt.GetState() == status.None {
			return customapm.TraceError(
				ctx,
				errorcatalog.Catalog.MustGet(errorcatalog.ErrForwardMissingQors),
				fsm.GetLogger(),
				fsm.counterForwardFailed,
			)
		}

		fsm.State = opt.GetState()

		// Load previous question.
		var prevQst question.Question

		if qst.PreviousQuestionID != "" {
			prevQst, _ = fsm.Questionnaire.Questions.Get(qst.PreviousQuestionID)
		}

		// Emit the state of the machine.
		fsm.Emit(prevQst, qst)
	}

	// Observability: metrics.
	fsm.counterForward.Add(1)

	return nil
}

// Load the FSM up to state of the event.
func Load(ctx context.Context, fsm *FiniteStateMachine, e event.Event) *FiniteStateMachine {
	fsm.Answers = e.Answers
	fsm.CurrentAnswer = e.CurrentAnswer
	fsm.CurrentQuestion = e.CurrentQuestion
	fsm.CurrentQuestionID = e.CurrentQuestion.GetID()
	fsm.CurrentQuestionIndex = e.CurrentQuestionIndex
	fsm.PreviousQuestion = e.PreviousQuestion
	fsm.PreviousQuestionID = e.PreviousQuestion.GetID()
	fsm.Questionnaire = e.Questionnaire
	fsm.State = e.State
	fsm.TotalAnswers = e.TotalAnswers
	fsm.TotalQuestions = e.TotalQuestions
	fsm.UserID = e.UserID

	return fsm
}

//////
// Factory.
//////

// New creates a new machine status.
//
// TODO: Also supports Channel.
func New(
	ctx context.Context,
	userID string,
	q questionnaire.Questionnaire,
	cb Callback,
) (*FiniteStateMachine, error) {
	// Storage's individual logger.
	logger := logging.Get().New(Name).SetTags(Type, Name)

	name := shared.RemoveSpacesAndToLower(q.Title)

	f := &FiniteStateMachine{
		Answers:       safeorderedmap.New[answer.Answer](),
		Questionnaire: q,
		Logger:        logger,
		State:         status.Initialized,
		UserID:        userID,

		callback: cb,

		counterBackward:            metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Runnning+".backward", DefaultMetricCounterLabel)),
		counterCompleted:           metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Completed, DefaultMetricCounterLabel)),
		counterDone:                metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Done, DefaultMetricCounterLabel)),
		counterEmitted:             metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Emitted, DefaultMetricCounterLabel)),
		counterForward:             metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Runnning+".forward", DefaultMetricCounterLabel)),
		counterForwardFailed:       metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Runnning+".forward"+"."+status.Failed, DefaultMetricCounterLabel)),
		counterInitialized:         metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Initialized, DefaultMetricCounterLabel)),
		counterInstantiationFailed: metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Instantiated+"."+status.Failed, DefaultMetricCounterLabel)),
		counterJump:                metrics.NewInt(fmt.Sprintf("%s.%s.%s.%s", Type, name, status.Runnning+".jump", DefaultMetricCounterLabel)),
	}

	// Validate the storage.
	if err := validation.Validate(f); err != nil {
		return nil, customapm.TraceError(ctx, err, logger, f.counterInstantiationFailed)
	}

	f.GetLogger().PrintlnWithOptions(level.Debug, status.Created.String())

	return f, nil
}
