package llmguard

// UnsafeDevelopmentEvent exposes raw diagnostic material for local development only.
// It may include original input, operation output, and finding metadata. TokenSet
// mappings are never included. UNSAFE FOR PRODUCTION — do not enable outside
// controlled development environments.
type UnsafeDevelopmentEvent struct {
	Operation Operation
	Outcome   Outcome
	Input     string
	Output    string
	Findings  []Finding
}

// UnsafeDevelopmentObserver receives explicit development-only diagnostics.
// Implementations must be concurrency-safe. UNSAFE FOR PRODUCTION.
type UnsafeDevelopmentObserver interface {
	ObserveUnsafe(UnsafeDevelopmentEvent)
}

// UnsafeDevelopmentObserverFunc adapts a function to UnsafeDevelopmentObserver.
// UNSAFE FOR PRODUCTION.
type UnsafeDevelopmentObserverFunc func(UnsafeDevelopmentEvent)

// ObserveUnsafe implements UnsafeDevelopmentObserver.
func (f UnsafeDevelopmentObserverFunc) ObserveUnsafe(event UnsafeDevelopmentEvent) {
	if f != nil {
		f(event)
	}
}

type unsafeDevelopmentObserverOption struct {
	observer UnsafeDevelopmentObserver
}

func (o unsafeDevelopmentObserverOption) apply(cfg *guardConfig) error {
	if isNilInterfaceValue(o.observer) {
		return newInvalidConfigError("unsafe development observer is nil")
	}
	if cfg.unsafeDevelopmentObserverSet {
		return newInvalidConfigError("duplicate unsafe development observer")
	}
	cfg.unsafeDevelopmentObserverSet = true
	cfg.unsafeDevelopmentObserver = o.observer
	return nil
}

// WithUnsafeDevelopmentObserver configures raw development diagnostics that may
// leak sensitive text and finding spans. This option is independent from
// WithObserver and never activates from safe observer configuration alone.
// UNSAFE FOR PRODUCTION.
func WithUnsafeDevelopmentObserver(observer UnsafeDevelopmentObserver) Option {
	return unsafeDevelopmentObserverOption{observer: observer}
}

func copyFindings(findings []Finding) []Finding {
	if len(findings) == 0 {
		return nil
	}
	out := make([]Finding, len(findings))
	copy(out, findings)
	return out
}

func (g *Guard) publishUnsafeEvent(event UnsafeDevelopmentEvent) {
	if g.unsafeDevelopmentObserver == nil {
		return
	}
	event.Findings = copyFindings(event.Findings)
	g.unsafeDevelopmentObserver.ObserveUnsafe(event)
}
