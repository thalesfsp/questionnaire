package metrics

import (
	"expvar"
	"fmt"
	"os"

	"github.com/thalesfsp/questionnaire/internal/logging"
)

// NewInt creates and initializes a new expvar.Int.
func NewInt(name string) *expvar.Int {
	prefix := os.Getenv("QUESTIONNAIRE_METRICS_PREFIX")

	if prefix == "" {
		logging.Get().Warnln("QUESTIONNAIRE_METRICS_PREFIX is not set. Using default (questionnaire).")

		prefix = "questionnaire"
	}

	counter := expvar.NewInt(
		fmt.Sprintf(
			"%s.%s",
			prefix,
			name,
		),
	)

	counter.Set(0)

	return counter
}
