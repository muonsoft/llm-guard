package llmguard

import "context"

// Detector finds entities in text. Implementations must return a stable non-empty
// Name and must not place sensitive data in the name string.
//
// A detector registered in a Guard must be safe for concurrent calls: one immutable
// Guard may invoke Detect from multiple goroutines at the same time.
type Detector interface {
	Name() string
	Detect(ctx context.Context, text string) ([]Finding, error)
}
