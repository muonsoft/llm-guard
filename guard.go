package llmguard

import (
	"context"
	"crypto/rand"
	"io"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muonsoft/errors"
)

// Option configures Guard construction.
type Option interface {
	apply(*guardConfig) error
}

type detectorOption struct {
	detector Detector
}

func (o detectorOption) apply(cfg *guardConfig) error {
	if isNilDetector(o.detector) {
		return newInvalidConfigError("detector is nil")
	}
	name := o.detector.Name()
	if err := validateDetectorName(name); err != nil {
		return err
	}
	for _, existing := range cfg.entries {
		if existing.name == name {
			return newInvalidConfigError("duplicate detector name")
		}
	}
	cfg.entries = append(cfg.entries, detectorEntry{
		name:     name,
		detector: o.detector,
	})
	return nil
}

// WithDetector registers a custom detector. Detector names must be unique.
func WithDetector(detector Detector) Option {
	return detectorOption{detector: detector}
}

type guardConfig struct {
	entries                      []detectorEntry
	randomSource                 io.Reader
	secretAction                 Action
	secretActionSet              bool
	entityActions                map[EntityType]Action
	observer                     Observer
	observerSet                  bool
	unsafeDevelopmentObserver    UnsafeDevelopmentObserver
	unsafeDevelopmentObserverSet bool
}

type detectorEntry struct {
	name     string
	detector Detector
}

// Guard runs registered detectors concurrently and aggregates validated findings.
type Guard struct {
	detectors                 []detectorEntry
	randomSource              io.Reader
	randomMu                  sync.Mutex
	secretAction              Action
	entityActions             map[EntityType]Action
	observer                  Observer
	unsafeDevelopmentObserver UnsafeDevelopmentObserver
}

// New constructs an immutable Guard from the given options.
func New(options ...Option) (*Guard, error) {
	cfg := guardConfig{}
	for _, option := range options {
		if option == nil {
			return nil, newInvalidConfigError("option is nil")
		}
		if err := option.apply(&cfg); err != nil {
			return nil, err
		}
	}

	entries := make([]detectorEntry, len(cfg.entries))
	copy(entries, cfg.entries)

	randomSource := cfg.randomSource
	if randomSource == nil {
		randomSource = rand.Reader
	}

	secretAction := cfg.secretAction
	if secretAction == "" {
		secretAction = ActionBlock
	}

	observer := cfg.observer
	if observer == nil {
		observer = NoopObserver{}
	}

	return &Guard{
		detectors:                 entries,
		randomSource:              randomSource,
		secretAction:              secretAction,
		entityActions:             copyEntityActions(cfg.entityActions),
		observer:                  observer,
		unsafeDevelopmentObserver: cfg.unsafeDevelopmentObserver,
	}, nil
}

type indexedFinding struct {
	Finding
	regIndex   int
	localIndex int
}

type detectorSlot struct {
	regIndex      int
	err           error
	derivedCancel bool
	findings      []indexedFinding
}

// Detect runs all configured detectors on text and returns validated findings
// in a stable order. On any failure it returns nil findings and a safe error.
func (g *Guard) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	started := time.Now()
	inputBytes := len(text)
	var (
		outcome  Outcome
		findings []Finding
		err      error
	)
	defer func() {
		event := Event{
			Operation:    OperationDetect,
			Outcome:      outcome,
			InputBytes:   inputBytes,
			OutputBytes:  0,
			Duration:     time.Since(started),
			FindingCount: len(findings),
		}
		if outcome == OutcomeSuccess {
			event.EntityCounts = buildEntityCounts(findings)
		}
		g.publishSafeEvent(event)
		g.publishUnsafeEvent(UnsafeDevelopmentEvent{
			Operation: OperationDetect,
			Outcome:   outcome,
			Input:     text,
			Output:    "",
			Findings:  findings,
		})
	}()

	if err = validateInputText(text); err != nil {
		outcome = OutcomeError
		return nil, err
	}
	if len(g.detectors) == 0 {
		outcome = OutcomeSuccess
		return nil, nil
	}

	detectCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	slots := make([]detectorSlot, len(g.detectors))
	var derivedCancelFlag atomic.Bool
	var wg sync.WaitGroup

	for i, entry := range g.detectors {
		slots[i].regIndex = i
		wg.Add(1)
		go func(slot *detectorSlot, entry detectorEntry) {
			defer wg.Done()

			detectorFindings, detectErr := entry.detector.Detect(detectCtx, text)
			if detectErr != nil {
				slot.err = newDetectorError(entry.name, detectErr)
				if isContextCancellation(detectErr) {
					if derivedCancelFlag.Load() {
						slot.derivedCancel = true
					} else {
						derivedCancelFlag.Store(true)
						cancel()
					}
				} else {
					derivedCancelFlag.Store(true)
					cancel()
				}
				return
			}

			validated, validateErr := validateFindings(text, entry.name, detectorFindings)
			if validateErr != nil {
				slot.err = validateErr
				derivedCancelFlag.Store(true)
				cancel()
				return
			}

			for j := range validated {
				validated[j].regIndex = slot.regIndex
			}
			slot.findings = validated
		}(&slots[i], entry)
	}

	wg.Wait()

	if err = ctx.Err(); err != nil {
		outcome = OutcomeError
		return nil, err
	}

	if err = selectSlotError(slots); err != nil {
		outcome = OutcomeError
		return nil, err
	}

	findings = aggregateFindings(slots)
	outcome = OutcomeSuccess
	return findings, nil
}

func selectSlotError(slots []detectorSlot) error {
	var chosen error
	chosenIndex := -1

	for _, slot := range slots {
		if slot.err == nil || slot.derivedCancel {
			continue
		}
		if chosen == nil || slot.regIndex < chosenIndex {
			chosen = slot.err
			chosenIndex = slot.regIndex
		}
	}

	return chosen
}

func isContextCancellation(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func aggregateFindings(slots []detectorSlot) []Finding {
	total := 0
	for _, slot := range slots {
		total += len(slot.findings)
	}
	if total == 0 {
		return nil
	}

	merged := make([]indexedFinding, 0, total)
	for _, slot := range slots {
		merged = append(merged, slot.findings...)
	}

	sort.SliceStable(merged, func(i, j int) bool {
		a, b := merged[i], merged[j]
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		if a.End != b.End {
			return a.End < b.End
		}
		if a.Entity != b.Entity {
			return a.Entity < b.Entity
		}
		if a.Detector != b.Detector {
			return a.Detector < b.Detector
		}
		if a.Confidence != b.Confidence {
			return a.Confidence > b.Confidence
		}
		if a.regIndex != b.regIndex {
			return a.regIndex < b.regIndex
		}
		return a.localIndex < b.localIndex
	})

	out := make([]Finding, len(merged))
	for i, f := range merged {
		out[i] = f.Finding
	}
	return out
}

func isNilDetector(detector Detector) bool {
	return isNilInterfaceValue(detector)
}

func isNilReader(reader io.Reader) bool {
	return isNilInterfaceValue(reader)
}

func isNilInterfaceValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
