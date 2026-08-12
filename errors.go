package llmguard

import (
	"fmt"

	"github.com/muonsoft/errors"
)

var (
	// ErrInvalidConfig indicates Guard construction received an invalid option.
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrInvalidText indicates Detect received text that is not valid UTF-8.
	ErrInvalidText = errors.New("invalid text")
	// ErrInvalidFinding indicates a detector returned a finding that failed validation.
	ErrInvalidFinding = errors.New("invalid finding")
	// ErrDetector indicates a detector returned an error during Detect.
	ErrDetector = errors.New("detector error")
	// ErrInvalidTokenSet indicates Restore received a nil or unusable TokenSet.
	ErrInvalidTokenSet = errors.New("invalid token set")
	// ErrNamespaceSource indicates namespace entropy could not be read.
	ErrNamespaceSource = errors.New("namespace source error")
	// ErrNamespaceCollision indicates no collision-free namespace was found.
	ErrNamespaceCollision = errors.New("namespace collision")
	// ErrBlocked indicates Mask was aborted by a block policy action.
	ErrBlocked = errors.New("blocked")
)

// InvalidConfigError describes why Guard construction failed.
type InvalidConfigError struct {
	Reason string
}

func (e *InvalidConfigError) Error() string {
	return fmt.Sprintf("invalid configuration: %s", e.Reason)
}

func (e *InvalidConfigError) Is(target error) bool {
	return errors.Is(ErrInvalidConfig, target)
}

func newInvalidConfigError(reason string) error {
	return errors.Wrap(&InvalidConfigError{Reason: reason}, errors.SkipCaller())
}

// InvalidFindingError reports which detector field failed validation without
// including the input text or matched substring.
type InvalidFindingError struct {
	Detector string
	Entity   EntityType
	Field    string
}

func (e *InvalidFindingError) Error() string {
	entity := string(e.Entity)
	if entity == "" {
		entity = "<empty>"
	}
	return fmt.Sprintf("invalid finding from detector %q for entity %q: %s", e.Detector, entity, e.Field)
}

func (e *InvalidFindingError) Is(target error) bool {
	return errors.Is(ErrInvalidFinding, target)
}

func newInvalidFindingError(detector string, entity EntityType, field string) error {
	return errors.Wrap(&InvalidFindingError{
		Detector: detector,
		Entity:   entity,
		Field:    field,
	}, errors.SkipCaller())
}

// DetectorError wraps a detector failure with a safe public message. The original
// cause remains available through Unwrap for errors.Is and errors.As.
type DetectorError struct {
	Detector string
	cause    error
}

func (e *DetectorError) Error() string {
	return fmt.Sprintf("detector %q failed", e.Detector)
}

func (e *DetectorError) Unwrap() error {
	return e.cause
}

func (e *DetectorError) Is(target error) bool {
	return errors.Is(ErrDetector, target)
}

func newDetectorError(detector string, cause error) error {
	return errors.Wrap(&DetectorError{
		Detector: detector,
		cause:    cause,
	}, errors.SkipCaller())
}

// InvalidTokenSetError reports why a TokenSet was rejected without exposing
// token, namespace, or mapping content.
type InvalidTokenSetError struct {
	Reason string
}

func (e *InvalidTokenSetError) Error() string {
	return fmt.Sprintf("invalid token set: %s", e.Reason)
}

func (e *InvalidTokenSetError) Is(target error) bool {
	return errors.Is(ErrInvalidTokenSet, target)
}

func newInvalidTokenSetError(reason string) error {
	return errors.Wrap(&InvalidTokenSetError{Reason: reason}, errors.SkipCaller())
}

// NamespaceSourceError wraps a failure to read namespace entropy.
type NamespaceSourceError struct {
	cause error
}

func (e *NamespaceSourceError) Error() string {
	return "namespace source error"
}

func (e *NamespaceSourceError) Unwrap() error {
	return e.cause
}

func (e *NamespaceSourceError) Is(target error) bool {
	return errors.Is(ErrNamespaceSource, target)
}

func newNamespaceSourceError(cause error) error {
	return errors.Wrap(&NamespaceSourceError{cause: cause}, errors.SkipCaller())
}

// NamespaceCollisionError reports that no collision-free namespace was found.
type NamespaceCollisionError struct{}

func (e *NamespaceCollisionError) Error() string {
	return "namespace collision"
}

func (e *NamespaceCollisionError) Is(target error) bool {
	return errors.Is(ErrNamespaceCollision, target)
}

func newNamespaceCollisionError() error {
	return errors.Wrap(&NamespaceCollisionError{}, errors.SkipCaller())
}

// BlockError reports that Mask was aborted by policy without exposing input,
// spans, entities, or occurrence metadata.
type BlockError struct{}

func (e *BlockError) Error() string {
	return "operation blocked by policy"
}

func (e *BlockError) Is(target error) bool {
	return errors.Is(ErrBlocked, target)
}

func newBlockError() error {
	return errors.Wrap(&BlockError{}, errors.SkipCaller())
}
