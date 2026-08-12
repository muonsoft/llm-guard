package llmguard

import (
	"context"

	"github.com/muonsoft/llm-guard/internal/nlp"
)

const addressDetectorName = "address"

type addressDetector struct{}

// NewAddressDetector returns an immutable built-in ADDRESS detector for conservative
// compositional Russian addresses. Register it with WithDetector when constructing
// a Guard.
func NewAddressDetector() Detector {
	return addressDetector{}
}

func (addressDetector) Name() string {
	return addressDetectorName
}

func (addressDetector) Detect(ctx context.Context, text string) ([]Finding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	spans, err := nlp.DetectAddressSpans(ctx, text)
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
			Entity:     EntityAddress,
			Start:      span.Start,
			End:        span.End,
			Confidence: 0.84,
			Detector:   addressDetectorName,
		})
	}
	return findings, nil
}
