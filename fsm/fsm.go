package fsm

import (
	"context"
	"expvar"
	"fmt"

	"github.com/thalesfsp/customerror"
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
// Var, const, and types.
//////

// Type is the type of the entity regarding the framework. It is used to for
// example, to identify the entity in the logs, metrics, and for tracing.
const (
	DefaultMetricCounterLabel = "counter"
	Name                      = "fsm"
	Type                      = "FiniteStateMachine"
)

// FiniteStateMachine is the Finite State Machine for the Questionnaire.
type FiniteStateMachine struct {
	// Answers is the list of answers.
	Answers *answer.Map `json:"answers"`

	// CurrentQuestionID is the ID of the current question.
	CurrentQuestionID string `json:"currentQuestionID"`

	// Journal is the list of events.
	Journal []event.Event `json:"journal"`

	// Logger.
	Logger sypl.ISypl `json:"-" validate:"required"`

	// PreviousQuestionID is the ID of the previous question.
	PreviousQuestionID string `json:"previousQuestionID"`

	// Questionnaire is the questionnaire.
	Questionnaire questionnaire.Questionnaire `json:"Questionnaire" validate:"required"`

	// State is the current state of the questionnaire.
	State status.Status `json:"state" validate:"required"`

	// UserID is the ID of the user.
	UserID string `json:"userID" validate:"required"`

	// Callback is the function that is called every time the state of the
	// questionnaire changes. For example: save the state to the database.
	callback event.Callback `json:"-"`

	// Metrics.
	counterBackward            *expvar.Int `json:"-" validate:"required,gte=0"`
	counterCompleted           *expvar.Int `json:"-" validate:"required,gte=0"`
	counterDone                *expvar.Int `json:"-" validate:"required,gte=0"`
	counterEmitted             *expvar.Int `json:"-" validate:"required,gte=0"`
	counterForward             *expvar.Int `json:"-" validate:"required,gte=0"`
	counterForwardFailed       *expvar.Int `json:"-" validate:"required,gte=0"`
	counterInitialized         *expvar.Int `json:"-" validate:"required,gte=0"`
	counterInstantiationFailed *expvar.Int `json:"-" validate:"required,gte=0"`
	counterJump                *expvar.Int `json:"-" validate:"required,gte=0"`
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
	if fsm.PreviousQuestionID == "" || fsm.State == status.Initialized {
		return fsm
	}

	// Load the current question - just to store reference.
	currentQst := fsm.Questionnaire.Questions.Load(fsm.CurrentQuestionID)

	// Load the answer based on the previous question ID. This is what matters.
	finalID := fsm.PreviousQuestionID

	// Allows to jump to a specific question.
	if id != "" {
		finalID = id
	}

	aswr := fsm.Answers.Load(finalID)

	// Do nothing if there's no answer.
	if aswr == nil {
		return fsm
	}

	// Load the answered question.
	answeredQst := aswr.GetQuestion()

	// Set the current question ID to the loaded question.
	fsm.CurrentQuestionID = answeredQst.ID

	// Set the previous question ID to the loaded question's previous question.
	fsm.PreviousQuestionID = answeredQst.PreviousQuestionID

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

// Create the answer based on the option.
func (fsm *FiniteStateMachine) optionToAnswer(
	ctx context.Context,
	qst question.Question,
	opt option.IOption,
) (answer.IAnswer, error) {
	var aswr answer.IAnswer

	switch opt := opt.(type) {
	case option.Option[bool]:
		aswr = answer.MustNew[bool](qst, opt)

	case option.Option[float64]:
		aswr = answer.MustNew[float64](qst, opt)

	case option.Option[float32]:
		aswr = answer.MustNew[float32](qst, opt)

	case option.Option[int]:
		aswr = answer.MustNew[int](qst, opt)

	case option.Option[string]:
		aswr = answer.MustNew[string](qst, opt)

	case option.Option[[]bool]:
		aswr = answer.MustNew[[]bool](qst, opt)

	case option.Option[[]float64]:
		aswr = answer.MustNew[[]float64](qst, opt)

	case option.Option[[]float32]:
		aswr = answer.MustNew[[]float32](qst, opt)

	case option.Option[[]int]:
		aswr = answer.MustNew[[]int](qst, opt)

	case option.Option[[]string]:
		aswr = answer.MustNew[[]string](qst, opt)
	default:
		return nil, customapm.TraceError(
			ctx,
			errorcatalog.Catalog.MustGet(errorcatalog.ErrAnswerOptionType, customerror.WithField("option", opt)),
			fsm.GetLogger(),
			fsm.counterForwardFailed,
		)
	}

	return aswr, nil
}

//////
// Special state machine methods.
//////

// Start the state machine.
func (fsm *FiniteStateMachine) Start() *FiniteStateMachine {
	// Transition to the answering status.
	fsm.State = status.Runnning

	// Load the first question.
	qst := fsm.Questionnaire.Questions.LoadByIndex(0)

	// Set the current question ID.
	fsm.CurrentQuestionID = qst.ID

	// Emit the state of the machine.
	fsm.Emit(question.Question{}, qst)

	// Observability: metrics.
	fsm.counterInitialized.Add(1)

	return fsm
}

// Forward the current question with the given option.
//
//nolint:nestif
func (fsm *FiniteStateMachine) Forward(ctx context.Context, opt option.IOption) error {
	// Ensure's the proper state is set.
	fsm.State = status.Runnning

	//////
	// Determine the current question.
	//////

	// Retrieves the current question from the Questionnaire.
	qst := fsm.Questionnaire.Questions.Load(fsm.CurrentQuestionID)

	fsm.PreviousQuestionID = qst.GetID()

	//////
	// Deal with answer.
	//
	// NOTE: The following are all the possible types of answers. Update
	// accordingly.
	//////

	aswr, err := fsm.optionToAnswer(ctx, qst, opt)
	if err != nil {
		return err
	}

	// Add answer to the list.
	fsm.Answers.Store(aswr.GetID(), aswr)

	// Run the answer function.
	if opt.GetAnswerFunc() != nil {
		if err := opt.GetAnswerFunc()(option.Answer{
			ID:             aswr.GetID(),
			QuestionID:     qst.GetID(),
			QuestionLabel:  qst.Label,
			QuestionWeight: qst.Meta.Weight,
			OptionID:       opt.GetID(),
			OptionLabel:    opt.GetLabel(),
			OptionValue:    opt.GetValue(),
			OptionWeight:   opt.GetWeight(),
		}); err != nil {
			return customapm.TraceError(
				ctx,
				err,
				fsm.GetLogger(),
				fsm.counterForwardFailed,
			)
		}
	}

	// If all questions have been answered, transition to the completed status.
	if fsm.Answers.Size() == fsm.Questionnaire.Questions.Size() {
		fsm.State = status.Completed
	}

	//////
	// Deal with the next question.
	//////

	// Evaluate the next question ID.
	nextQstID := aswr.GetOption().NextQuestionID()

	if nextQstID != "" {
		// Loads the question from the Questionnaire.
		nextQst := fsm.Questionnaire.Questions.Load(nextQstID)

		//////
		// Deal with setting the previous question.
		//////

		// Sets the previous question ID.
		if nextQst.PreviousQuestionID == "" {
			nextQst.PreviousQuestionID = qst.GetID()
		}

		// Update questionnaire's questions with the updated one persisting the
		// change.
		fsm.Questionnaire.Questions.Store(nextQst.GetID(), nextQst)

		//////
		// Deal with settings the current question ID, and determining the status.
		//////

		// Update the current question ID.
		fsm.CurrentQuestionID = nextQst.GetID()

		// Optionally, set the state based on the option (answer).
		if aswr.GetOption().GetState() != status.None {
			fsm.State = aswr.GetOption().GetState()
		}

		// Emit the state of the machine.
		fsm.Emit(qst, nextQst)
	} else {
		if aswr.GetOption().GetState() == status.None {
			return customapm.TraceError(
				ctx,
				errorcatalog.Catalog.MustGet(errorcatalog.ErrForwardMissingQors),
				fsm.GetLogger(),
				fsm.counterForwardFailed,
			)
		}

		fsm.State = aswr.GetOption().GetState()

		// Load previous question.
		var prevQst question.Question

		if qst.PreviousQuestionID != "" {
			prevQst = fsm.Questionnaire.Questions.Load(qst.PreviousQuestionID)
		}

		// Emit the state of the machine.
		fsm.Emit(prevQst, qst)
	}

	// Observability: metrics.
	fsm.counterForward.Add(1)

	return nil
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
	e := event.Event{
		CurrentQuestion:  currentQst,
		State:            fsm.State,
		TotalAnswers:     fsm.Answers.Size(),
		TotalQuestions:   fsm.Questionnaire.Questions.Size(),
		UserID:           fsm.UserID,
		PreviousQuestion: prevQst,
		Answers:          fsm.Answers,
		Questionnaire:    fsm.Questionnaire,
	}

	// Emit the state of the machine.
	fsm.callback(e)

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
	return fsm.Emit(
		fsm.Questionnaire.Questions.Load(fsm.PreviousQuestionID),
		fsm.Questionnaire.Questions.Load(fsm.CurrentQuestionID),
	)
}

//////
// Factory.
//////

// New creates a new machine status.
func New(
	ctx context.Context,
	userID string,
	q questionnaire.Questionnaire,
	cb event.Callback,
) (*FiniteStateMachine, error) {
	// Storage's individual logger.
	logger := logging.Get().New(Name).SetTags(Type, Name)

	name := shared.RemoveSpacesAndToLower(q.Title)

	f := &FiniteStateMachine{
		Answers:       answer.NewMap(),
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
