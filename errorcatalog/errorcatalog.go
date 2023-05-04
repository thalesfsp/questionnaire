package errorcatalog

import "github.com/thalesfsp/customerror"

const (
	ErrAnswerOptionRequired = "ERR_ANSWER_OPTION_REQUIRED"
	ErrAnswerOptionType     = "ERR_ANSWER_OPTION_TYPE"
	ErrForwardMissingQors   = "ERR_FORWARD_MISSING_QORS"
)

// Catalog of errors.
var Catalog = customerror.
	MustNewCatalog("questionnaire").
	MustSet(ErrAnswerOptionRequired, "Question's answer is required").
	MustSet(ErrAnswerOptionType, "Answer's option type is invalid").
	MustSet(ErrForwardMissingQors, "Missing setting the question ID or the state")
