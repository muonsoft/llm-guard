package evaluation

import (
	"context"

	"github.com/muonsoft/llm-guard"
)

// DetectResolve runs guard.Detect followed by llmguard.Resolve on input text.
func DetectResolve(ctx context.Context, guard *llmguard.Guard, input string) ([]llmguard.Finding, error) {
	findings, err := guard.Detect(ctx, input)
	if err != nil {
		return nil, err
	}
	return llmguard.Resolve(input, findings)
}
