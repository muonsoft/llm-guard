package llmguard_test

import (
	"context"

	"github.com/muonsoft/llm-guard"
)

type atSignDetector struct{}

func (atSignDetector) Name() string { return "at-sign" }

func (atSignDetector) Detect(_ context.Context, text string) ([]llmguard.Finding, error) {
	for i := 0; i < len(text); i++ {
		if text[i] == '@' {
			return []llmguard.Finding{{
				Entity:     llmguard.EntityEmail,
				Start:      i,
				End:        i + 1,
				Confidence: 0.8,
			}}, nil
		}
	}
	return nil, nil
}

func ExampleNew_customDetector() {
	guard, err := llmguard.New(llmguard.WithDetector(atSignDetector{}))
	if err != nil {
		panic(err)
	}

	findings, err := guard.Detect(context.Background(), "contact a@example.com")
	if err != nil {
		panic(err)
	}

	for _, finding := range findings {
		_ = finding.Entity
		_ = finding.Start
		_ = finding.End
	}
	// Output:
}
