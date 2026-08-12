package llmguard

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// Operation identifies which Guard API produced an observer event.
type Operation string

const (
	OperationDetect  Operation = "detect"
	OperationMask    Operation = "mask"
	OperationRestore Operation = "restore"
)

// Outcome is a stable low-cardinality terminal result for an operation.
type Outcome string

const (
	OutcomeSuccess     Outcome = "success"
	OutcomeError       Outcome = "error"
	OutcomeBlocked     Outcome = "blocked"
	OutcomeRestoreMiss Outcome = "restore_miss"
)

// EntityCount reports how many findings or resolved spans belong to one entity.
type EntityCount struct {
	Entity EntityType
	Count  int
}

// ActionCount reports how many resolved findings received a policy action.
type ActionCount struct {
	Action Action
	Count  int
}

// Event is a safe terminal observability payload for one Detect, Mask, or Restore call.
// It never contains original, masked, or restored text, finding values, detector
// causes, placeholder material, or TokenSet mappings.
type Event struct {
	Operation     Operation
	Outcome       Outcome
	InputBytes    int
	OutputBytes   int
	Duration      time.Duration
	FindingCount  int
	RestoreMisses int
	EntityCounts  []EntityCount
	ActionCounts  []ActionCount
}

// String returns a safe single-line summary.
func (e Event) String() string {
	return fmt.Sprintf(
		"llmguard.Event{operation=%s outcome=%s input_bytes=%d output_bytes=%d duration=%s finding_count=%d restore_misses=%d entity_counts=%d action_counts=%d}",
		e.Operation, e.Outcome, e.InputBytes, e.OutputBytes, e.Duration, e.FindingCount, e.RestoreMisses, len(e.EntityCounts), len(e.ActionCounts),
	)
}

// GoString returns the same safe summary as String.
func (e Event) GoString() string {
	return e.String()
}

// MarshalJSON returns a deterministic JSON representation without sensitive values.
func (e Event) MarshalJSON() ([]byte, error) {
	type safeEntityCount struct {
		Entity string `json:"entity"`
		Count  int    `json:"count"`
	}
	type safeActionCount struct {
		Action string `json:"action"`
		Count  int    `json:"count"`
	}
	type safeEvent struct {
		Operation     string            `json:"operation"`
		Outcome       string            `json:"outcome"`
		InputBytes    int               `json:"input_bytes"`
		OutputBytes   int               `json:"output_bytes"`
		DurationNS    int64             `json:"duration_ns"`
		FindingCount  int               `json:"finding_count"`
		RestoreMisses int               `json:"restore_misses"`
		EntityCounts  []safeEntityCount `json:"entity_counts"`
		ActionCounts  []safeActionCount `json:"action_counts"`
	}
	payload := safeEvent{
		Operation:     string(e.Operation),
		Outcome:       string(e.Outcome),
		InputBytes:    e.InputBytes,
		OutputBytes:   e.OutputBytes,
		DurationNS:    e.Duration.Nanoseconds(),
		FindingCount:  e.FindingCount,
		RestoreMisses: e.RestoreMisses,
	}
	for _, count := range e.EntityCounts {
		payload.EntityCounts = append(payload.EntityCounts, safeEntityCount{
			Entity: string(observabilityEntity(count.Entity)),
			Count:  count.Count,
		})
	}
	for _, count := range e.ActionCounts {
		payload.ActionCounts = append(payload.ActionCounts, safeActionCount{
			Action: string(count.Action),
			Count:  count.Count,
		})
	}
	return json.Marshal(payload)
}

// Observer receives terminal safe events. Implementations must be concurrency-safe
// because Guard may invoke callbacks from concurrent operations.
type Observer interface {
	Observe(Event)
}

// ObserverFunc adapts a function to Observer.
type ObserverFunc func(Event)

// Observe implements Observer.
func (f ObserverFunc) Observe(event Event) {
	if f != nil {
		f(event)
	}
}

// NoopObserver discards all events.
type NoopObserver struct{}

// Observe implements Observer.
func (NoopObserver) Observe(Event) {}

type observerOption struct {
	observer Observer
}

func (o observerOption) apply(cfg *guardConfig) error {
	if isNilInterfaceValue(o.observer) {
		return newInvalidConfigError("observer is nil")
	}
	if cfg.observerSet {
		return newInvalidConfigError("duplicate observer")
	}
	cfg.observerSet = true
	cfg.observer = o.observer
	return nil
}

// WithObserver configures a safe production observer. Without this option Guard
// uses NoopObserver and emits no callbacks.
func WithObserver(observer Observer) Option {
	return observerOption{observer: observer}
}

func copyEntityCounts(counts []EntityCount) []EntityCount {
	if len(counts) == 0 {
		return nil
	}
	out := make([]EntityCount, len(counts))
	copy(out, counts)
	return out
}

func copyActionCounts(counts []ActionCount) []ActionCount {
	if len(counts) == 0 {
		return nil
	}
	out := make([]ActionCount, len(counts))
	copy(out, counts)
	return out
}

func buildEntityCounts(findings []Finding) []EntityCount {
	if len(findings) == 0 {
		return nil
	}
	counts := make(map[EntityType]int, len(findings))
	for _, finding := range findings {
		counts[observabilityEntity(finding.Entity)]++
	}
	out := make([]EntityCount, 0, len(counts))
	for entity, count := range counts {
		out = append(out, EntityCount{Entity: entity, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Entity < out[j].Entity
	})
	return out
}

func buildActionCounts(findings []Finding, actionFor func(EntityType) Action) []ActionCount {
	if len(findings) == 0 {
		return nil
	}
	counts := make(map[Action]int)
	for _, finding := range findings {
		counts[actionFor(finding.Entity)]++
	}
	out := make([]ActionCount, 0, len(counts))
	for action, count := range counts {
		out = append(out, ActionCount{Action: action, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Action < out[j].Action
	})
	return out
}

func (g *Guard) publishSafeEvent(event Event) {
	if g.observer == nil {
		return
	}
	event.EntityCounts = copyEntityCounts(event.EntityCounts)
	event.ActionCounts = copyActionCounts(event.ActionCounts)
	g.observer.Observe(event)
}
