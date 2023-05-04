package fsm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thalesfsp/questionnaire/event"
	"github.com/thalesfsp/questionnaire/internal/shared"
	"github.com/thalesfsp/questionnaire/option"
	"github.com/thalesfsp/questionnaire/question"
	"github.com/thalesfsp/questionnaire/questionnaire"
	"github.com/thalesfsp/questionnaire/types"
	"github.com/thalesfsp/status"
)

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

			// NOTE: There are a few options:
			// - Set the next question ID.
			// - Set the next question ID by a function.
			// - Set the state of the option.
			// - Set the postAnswer hook
			red := option.MustNew("Red", option.WithNextQuestionID(q2ID))
			blue := option.MustNew("Blue", option.WithNextQuestionFunc(func() string {
				return q3ID
			}))

			n0 := option.MustNew(0, option.WithNextQuestionID(q3ID), option.WithForwardFunc(func(option.Answer) error {
				// Do something here, example, if user did surf for 10 years,
				// recommend sunscreen with SPF 50, and a skin cancer checkup.
				return nil
			}))
			n1 := option.MustNew(1, option.WithState(status.Completed))

			bTrue := option.MustNew(true, option.WithNextQuestionID(q2ID))
			bFalse := option.MustNew(false, option.WithNextQuestionID(q1ID))

			//////
			// Questions.
			//////

			q1 := question.MustNew(
				q1ID, "What's your favorite color?", types.Text,
				// Simply add as many options as you want, easy.
				question.WithOption(red, blue),
			)

			q2 := question.MustNew(
				q2ID, "How many years of programming experience do you have?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(n0, n1),
			)

			q3 := question.MustNew(
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

			//////
			// User State Machine.
			//////

			fsm, err := New(ctx, "12345", *q, func(e event.Event) {
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
					t.Fatal(err)
				}

				// Add break line.
				b = append(b, []byte("\n")...)

				file, err := os.OpenFile("status.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				if err != nil {
					t.Fatal(err)
				}

				defer file.Close()

				if _, err := file.Write(b); err != nil {
					t.Fatal(err)
				}

				// NOTE: This is an example of what to do when the questionnaire
				// is finished.
				//
				// if e.State == status.Done {
				// Send email, notify SMS, do whatever you want.
				// }
			})

			assert.NoError(t, err)

			// Start the machine status.
			fsm.Start()

			//////
			// The next steps simulates the user answering the questions.
			//
			// Answer the questions.
			//////

			// Answers Q1: "wyfc"
			err = fsm.Forward(ctx, blue) // Goes to Q3: "ayfgpl"
			assert.NoError(t, err)

			// Answers Q3: "ayfgpl"
			err = fsm.Forward(ctx, bTrue) // Goes to Q2: "hmyopedyh"
			assert.NoError(t, err)

			// Answers Q2: "hmyopedyh"
			err = fsm.Forward(ctx, n0) // Goes to Q3: "ayfgpl"
			assert.NoError(t, err)

			// Current Q3: "ayfgpl"
			// Should load Q2: "hmyopedyh"
			fsm.Backward()

			// Answers Q2: "hmyopedyh"
			err = fsm.Forward(ctx, n1) // Goes to nowhere, options set the state to "Finished".
			assert.NoError(t, err)

			// State should be "Finished".
			assert.Equal(t, status.Completed, fsm.GetState())

			// Current Q2: "hmyopedyh"
			// Should load Q1: "wyfc"
			fsm.Jump(q1ID)

			// State should be "Answering".
			assert.Equal(t, status.Runnning, fsm.GetState())

			// Answers Q1: "wyfc"
			err = fsm.Forward(ctx, red) // Goes to Q2: "hmyopedyh"
			assert.NoError(t, err)

			// Finish the questionnaire, now by calling Finish().
			fsm.Done()

			// State should be "Done".
			assert.Equal(t, status.Done, fsm.GetState())

			//////
			// This demonstrates the save feature. It can be called anytime.
			//////

			b, err := shared.Marshal(fsm.Dump())
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
			b2, err := shared.Marshal(fsm.GetJournal())
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

			q1 := question.MustNew(
				"1", "String?", types.Text,
				// Simply add as many options as you want, easy.
				question.WithOption(red),
			)

			q2 := question.MustNew(
				"2", "Int?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(n1),
			)

			q3 := question.MustNew(
				"3", "Bool?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(bTrue),
			)

			q4 := question.MustNew(
				"4", "Float32?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(f1),
			)

			q5 := question.MustNew(
				"5", "Float64?", types.SingleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(f2),
			)

			q6 := question.MustNew(
				"6", "Slice of strings?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(ss),
			)

			q7 := question.MustNew(
				"7", "Slice of ints?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(si),
			)

			q8 := question.MustNew(
				"8", "Slice of bool?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(sb),
			)

			q9 := question.MustNew(
				"9", "Slice of float32?", types.MultipleSelect,
				// Simply add as many options as you want, easy.
				question.WithOption(sf1),
			)

			q10 := question.MustNew(
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

			fsm, err := New(ctx, "12345", *q, func(e event.Event) {
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

			err = fsm.Forward(ctx, red)
			assert.NoError(t, err)
			assert.Equal(t, "2", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, n1)
			assert.NoError(t, err)
			assert.Equal(t, "3", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, bTrue)
			assert.NoError(t, err)
			assert.Equal(t, "4", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, f1)
			assert.NoError(t, err)
			assert.Equal(t, "5", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, f2)
			assert.NoError(t, err)
			assert.Equal(t, "6", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, ss)
			assert.NoError(t, err)
			assert.Equal(t, "7", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, si)
			assert.NoError(t, err)
			assert.Equal(t, "8", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, sb)
			assert.NoError(t, err)
			assert.Equal(t, "9", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, sf1)
			assert.NoError(t, err)
			assert.Equal(t, "10", fsm.CurrentQuestionID)

			err = fsm.Forward(ctx, sf2)
			assert.NoError(t, err)
			assert.Equal(t, "10", fsm.CurrentQuestionID)
		})
	}
}
