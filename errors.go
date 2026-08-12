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
