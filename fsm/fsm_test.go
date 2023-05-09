package fsm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thalesfsp/questionnaire/answer"
	"github.com/thalesfsp/questionnaire/event"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/question"
	"github.com/thalesfsp/questionnaire/questionnaire"
	"github.com/thalesfsp/questionnaire/types"
	"github.com/thalesfsp/status"
)

var cb = func(e event.Event, journal []event.Event) {
	// This is the callback that will be called every time the
	// state machine changes its status. Do whatever you want with
	// the event, the obvious thing is TO SAVE IT (STATE) SOMEWHERE,
	// or you can marshal it to JSON and send to a message broker,
	// it's up to you!
	//
	// NOTE: This is a snapshot of the state of the machine which
	// contains the questionnaire, questions, etc. THIS (snapshot)
	// IS included ON PUPROSE for AUDIT. People's choice should not
	// be altered/changed if the questionnaire, its questions, or
	// its options are changed. Once a questionnaire is started,
	// IT SHOULD BE IMMUTABLE, should not be TAMPERED!
	//
	// NOTE: Proper error handling is required as it's a callback.
	// Observability: APM, trace, log, metrics, etc!

	b, err := shared.Marshal(e)
	if err != nil {
		panic(err)
	}

	// Add break line.
	b = append(b, []byte("\n")...)

	file, err := os.OpenFile("status.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	if _, err := file.Write(b); err != nil {
		panic(err)
	}

	// NOTE: This is an example of what to do when the questionnaire
	// is finished.
	//
	// if e.State == status.Done {
	// Send email, notify SMS, do whatever you want.
	// }
}

func TestNew_branching(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Should work",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			//////
			// The next steps simulates creating the questionnaire.
			//////

			// Predefined question IDs. To be used further for branching.
			q1ID := "wyfc"
			q2ID := "hmyopedyh"
			q3ID := "ayfgpl"

			//////
			// Question's options.
			//////

			red := option.MustNew("Red", option.WithNextQuestionID(q2ID))
			blue := option.MustNew("Blue", option.WithNextQuestionFunc(func() string {
				return q3ID
			}))

			n0 := option.MustNew(0, option.WithNextQuestionID(q3ID))
			n1 := option.MustNew(1, option.WithState(status.Completed))

			bTrue := option.MustNew(true, option.WithNextQuestionID(q2ID))
			bFalse := option.MustNew(false, option.WithNextQuestionID(q1ID))

			//////
			// Questions.
			//////

			q1 := question.MustNew[string](
				q1ID, "What's your favorite color?", types.Text,
				// Simply add as many options as you want, easy.
				question.WithOption(red, blue),
			)

			q2 := question.MustNew[int](
				q2ID, "How many years of programming experience do you have?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(n0, n1),
			)

			q3 := question.MustNew[bool](
				q3ID, "Are you familiar with the Go programming language?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(bTrue, bFalse),
			)

			//////
			// Questionnaire.
			//////

			// Simply add as many questions as you want, easy.
			q, err := questionnaire.New("Simple Survey - 1", q1, q2, q3)
			if err != nil {
				t.Fatal(err)
			}

			// Convert q to JSON.
			bq, err := shared.Marshal(q)
			if err != nil {
				t.Fatal(err)
			}

			// Convert bq to questionnaire.
			var qq1 questionnaire.Questionnaire
			if err := shared.Unmarshal(bq, &qq1); err != nil {
				t.Fatal(err)
			}

			//////
			// User State Machine.
			//////

			fsm, err := New(ctx, "12345", *q, cb)
			assert.NoError(t, err)

			if fsm == nil {
				t.Fatal("fsm is nil")
			}

			// Start the machine status.
			fsm.Start()

			//////
			// The next steps simulates the user answering the questions.
			//
			// Answer the questions.
			//////

			// Answers Q1: "wyfc"
			err = Forward(ctx, fsm, blue) // Goes to Q3: "ayfgpl"
			assert.NoError(t, err)

			_, answerOptBlue, _ := fsm.Answers.Last()
			answerBlue := answer.GetOption[string](answerOptBlue)
			assert.Equal(t, blue.Value, answerBlue.Value)

			// Answers Q3: "ayfgpl"
			err = Forward(ctx, fsm, bTrue) // Goes to Q2: "hmyopedyh"
			assert.NoError(t, err)

			_, answerOptTrue, _ := fsm.Answers.Last()
			answerTrue := answer.GetOption[bool](answerOptTrue)
			assert.Equal(t, bTrue.Value, answerTrue.Value)

			// Answers Q2: "hmyopedyh"
			err = Forward(ctx, fsm, n0) // Goes to Q3: "ayfgpl"
			assert.NoError(t, err)

			_, answerOpt0, _ := fsm.Answers.Last()
			answer0 := answer.GetOption[int](answerOpt0)
			assert.Equal(t, n0.Value, answer0.Value)

			// Current Q3: "ayfgpl"
			// Should load Q2: "hmyopedyh"
			fsm.Backward()

			answerOpt02, _ := fsm.Answers.Get(q2ID)
			assert.Equal(t, 0, answer.GetOption[int](answerOpt02).Value)

			// Answers Q2: "hmyopedyh"
			err = Forward(ctx, fsm, n1) // Goes to nowhere, options set the state to "Finished".
			assert.NoError(t, err)

			answerOpt01, _ := fsm.Answers.Get(q2ID)
			assert.Equal(t, n1.Value, answer.GetOption[int](answerOpt01).Value)
			assert.Equal(t, status.Completed, fsm.GetState())

			//////
			// Simulate saving to the database.
			//////

			dump := fsm.Dump()
			assert.NotEmpty(t, dump)

			b, err := shared.Marshal(dump)
			assert.NoError(t, err)
			assert.NotEmpty(t, b)

			//////
			// Simulate loading from the database.
			//////

			var loadedDump event.Event
			_ = shared.Unmarshal(b, &loadedDump)

			fsm2, err2 := New(ctx, "12345", *q, cb)
			assert.NoError(t, err2)

			// Load the state machine from the dump.
			fsm2 = Load(ctx, fsm2, loadedDump)

			// Current Q2: "hmyopedyh"
			// Should load Q1: "wyfc"
			fsm2.Jump(q1ID)
			// fsm.Jump(q1ID)

			// State should be "Answering".
			assert.Equal(t, status.Runnning, fsm2.GetState())

			// Answers Q1: "wyfc"
			err = Forward(ctx, fsm2, red) // Goes to Q2: "hmyopedyh"
			assert.NoError(t, err)

			// Finish the questionnaire, now by calling Finish().
			fsm2.Done()

			// State should be "Done".
			assert.Equal(t, status.Done, fsm2.GetState())

			//////
			// This demonstrates the dump feature. It can be called anytime.
			//////

			b, err = shared.Marshal(fsm2.Dump())
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile("final.json", b, 0o600); err != nil {
				t.Fatal(err)
			}

			//////
			// This demonstrates the journal feature. It can be used to replay
			// the state machine. Yep, you can replay the state machine!
			//////

			// This demonstrates the save feature. It can be called anytime.
			b2, err := shared.Marshal(fsm2.GetJournal())
			if err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile("journal.json", b2, 0o600); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNew_testTypes(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Should work",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			//////
			// The next steps simulates creating the questionnaire.
			//////

			//////
			// Question's options.
			//////

			// String.
			red := option.MustNew("Red", option.WithNextQuestionID("2"))

			// Int.
			n1 := option.MustNew(1, option.WithNextQuestionID("3"))

			// Bool.
			bTrue := option.MustNew(true, option.WithNextQuestionID("4"))

			// Float32.
			f1 := option.MustNew(float32(1.1), option.WithNextQuestionID("5"))

			// Float64.
			f2 := option.MustNew(float64(1.1), option.WithNextQuestionID("6"))

			// Slice of strings.
			ss := option.MustNew([]string{"a", "b", "c"}, option.WithNextQuestionID("7"))

			// Slice of ints.
			si := option.MustNew([]int{1, 2, 3}, option.WithNextQuestionID("8"))

			// Slice of bool.
			sb := option.MustNew([]bool{true, false, true}, option.WithNextQuestionID("9"))

			// Slice of float32.
			sf1 := option.MustNew([]float32{1.1, 2.2, 3.3}, option.WithNextQuestionID("10"))

			// Slice of float64.
			sf2 := option.MustNew([]float64{1.1, 2.2, 3.3}, option.WithState(status.Completed))

			//////
			// Questions.
			//////

			q1 := question.MustNew[string](
				"1", "String?", types.Text,
				// Simply add as many options as you want, easy.
				question.WithOption(red),
			)

			q2 := question.MustNew[int](
				"2", "Int?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(n1),
			)

			q3 := question.MustNew[bool](
				"3", "Bool?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(bTrue),
			)

			q4 := question.MustNew[float32](
				"4", "Float32?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(f1),
			)

			q5 := question.MustNew[float64](
				"5", "Float64?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(f2),
			)

			q6 := question.MustNew[[]string](
				"6", "Slice of strings?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(ss),
			)

			q7 := question.MustNew[[]int](
				"7", "Slice of ints?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(si),
			)

			q8 := question.MustNew[[]bool](
				"8", "Slice of bool?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(sb),
			)

			q9 := question.MustNew[[]float32](
				"9", "Slice of float32?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(sf1),
			)

			q10 := question.MustNew[[]float64](
				"10", "Slice of float64?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(sf2),
			)

			//////
			// Questionnaire.
			//////

			q, err := questionnaire.New("Simple Survey - 2",
				// Simply add as many questions as you want, easy.
				q1, q2, q3, q4, q5, q6, q7, q8, q9, q10,
			)
			if err != nil {
				t.Fatal(err)
			}

			//////
			// User State Machine.
			//////

			fsm, err := New(ctx, "12345", *q, func(e event.Event, journal []event.Event) {
				// This is the callback that will be called every time the
				// state machine changes its status. Do whatever you want with
				// the event, the obvious thing is to save it (state) somewhere,
				// or you can marshal it to JSON and send to a message broker,
				// it's up to you!
				//
				// NOTE: Proper error handling is required as it's a callback.
				// Observability: APM, trace, log, metrics!
			})

			assert.NoError(t, err)
			if fsm == nil {
				t.Fatal("fsm is nil")
			}

			// Start the machine status.
			fsm.Start()

			//////
			// The next steps simulates the user answering the questions.
			//
			// Answer the questions.
			//////

			//////
			// Answer the questions.
			//////

			assert.Equal(t, "1", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, red)
			assert.NoError(t, err)
			assert.Equal(t, "2", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, n1)
			assert.NoError(t, err)
			assert.Equal(t, "3", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, bTrue)
			assert.NoError(t, err)
			assert.Equal(t, "4", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, f1)
			assert.NoError(t, err)
			assert.Equal(t, "5", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, f2)
			assert.NoError(t, err)
			assert.Equal(t, "6", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, ss)
			assert.NoError(t, err)
			assert.Equal(t, "7", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, si)
			assert.NoError(t, err)
			assert.Equal(t, "8", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, sb)
			assert.NoError(t, err)
			assert.Equal(t, "9", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, sf1)
			assert.NoError(t, err)
			assert.Equal(t, "10", fsm.CurrentQuestionID)

			err = Forward(ctx, fsm, sf2)
			assert.NoError(t, err)
			assert.Equal(t, "10", fsm.CurrentQuestionID)
		})
	}
}
