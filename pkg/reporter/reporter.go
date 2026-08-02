package reporter

import (
	"io"

	"github.com/archguard/archguard/internal/core"
)

// Reporter defines the standard interface for formatting and outputting scan results.
type Reporter interface {
	Report(w io.Writer, res *core.ScanResult) error
}
