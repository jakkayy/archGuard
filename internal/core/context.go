package core

import "context"

// ScanContext holds runtime context and target file list for scanning.
type ScanContext struct {
	Ctx        context.Context
	WorkingDir string
	Files      []string
	Options    map[string]any
}

// NewScanContext creates a new initialized ScanContext with a fallback background context.
func NewScanContext(ctx context.Context, workingDir string, files []string) *ScanContext {
	if ctx == nil {
		ctx = context.Background()
	}
	return &ScanContext{
		Ctx:        ctx,
		WorkingDir: workingDir,
		Files:      files,
		Options:    make(map[string]any),
	}
}
