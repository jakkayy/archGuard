package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/archguard/archguard/internal/core"
)

// JSONReporter formats scan results into indented JSON output.
type JSONReporter struct{}

// NewJSONReporter initializes a new JSONReporter instance.
func NewJSONReporter() *JSONReporter {
	return &JSONReporter{}
}

// Report writes formatted JSON scan results to the provided io.Writer.
func (j *JSONReporter) Report(w io.Writer, res *core.ScanResult) error {
	if res == nil {
		return fmt.Errorf("cannot format nil scan result")
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(res); err != nil {
		return fmt.Errorf("failed to encode scan result to json: %w", err)
	}

	return nil
}
