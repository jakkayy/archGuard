package reporter

import (
	"io"

	"github.com/jakkayy/archGuard/pkg/policy"
)

// Reporter defines the standard interface for formatting and outputting scan results.
type Reporter interface {
	Report(w io.Writer, res *policy.ScanResult) error
}
