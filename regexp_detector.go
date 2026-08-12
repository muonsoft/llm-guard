package llmguard

import (
	"context"
	"regexp"
)

// RegexDetectorConfig configures a compile-once custom regexp detector.
type RegexDetectorConfig struct {
	Name       string
	Entity     EntityType
	Pattern    string
	Confidence float64
}

// CustomRegexpDetector applies a caller-owned RE2 pattern as an immutable Detector.
// It must be constructed with NewCustomRegexpDetector; the zero value is not usable.
type CustomRegexpDetector struct {
	name       string
	entity     EntityType
	confidence float64
	pattern    *regexp.Regexp
}

// NewCustomRegexpDetector compiles config.Pattern once and returns a reusable detector.
func NewCustomRegexpDetector(config RegexDetectorConfig) (*CustomRegexpDetector, error) {
	if err := validateDetectorName(config.Name); err != nil {
		return nil, err
	}
	if config.Entity == "" {
		return nil, newInvalidConfigError("entity is empty")
	}
	if config.Pattern == "" {
		return nil, newInvalidConfigError("pattern is empty")
	}
	if !isValidConfidence(config.Confidence) {
		return nil, newInvalidConfigError("confidence out of range")
	}

	compiled, err := regexp.Compile(config.Pattern)
	if err != nil {
		return nil, newInvalidConfigError("invalid pattern")
	}

	return &CustomRegexpDetector{
		name:       config.Name,
		entity:     config.Entity,
		confidence: config.Confidence,
		pattern:    compiled,
	}, nil
}

func (d *CustomRegexpDetector) Name() string {
	return d.name
}

func (d *CustomRegexpDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := d.pattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	findings := make([]Finding, 0, len(matches))
	for _, loc := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		start, end := loc[0], loc[1]
		if start == end {
			continue
		}
		findings = append(findings, Finding{
			Entity:     d.entity,
			Start:      start,
			End:        end,
			Confidence: d.confidence,
		})
	}

	if len(findings) == 0 {
		return nil, nil
	}
	return findings, nil
}
