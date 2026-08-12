package llmguard

import (
	"context"

	"github.com/muonsoft/llm-guard/internal/nlp"
)

const personDetectorName = "person"

type personDetector struct{}

// NewPersonDetector returns an immutable built-in PERSON detector for conservative
// Russian full-name and initials sequences. Register it with WithDetector when
// constructing a Guard.
func NewPersonDetector() Detector {
	return personDetector{}
}

func (personDetector) Name() string {
	return personDetectorName
}

func (personDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	spans, err := nlp.DetectPersonSpans(ctx, text)
	if err != nil {
		return nil, err
	}
	if len(spans) == 0 {
		return nil, nil
	}

	findings := make([]Finding, 0, len(spans))
	for _, span := range spans {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		findings = append(findings, Finding{
			Entity:     EntityPerson,
			Start:      span.Start,
			End:        span.End,
			Confidence: 0.82,
			Detector:   personDetectorName,
		})
	}
	return findings, nil
}
