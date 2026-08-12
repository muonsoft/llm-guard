package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type keywordDetector struct {
	name    string
	keyword string
	entity  llmguard.EntityType
}

func (d *keywordDetector) Name() string { return d.name }

func (d *keywordDetector) Detect(_ context.Context, text string) ([]llmguard.Finding, error) {
	start := 0
	for start <= len(text)-len(d.keyword) {
		if text[start:start+len(d.keyword)] == d.keyword {
			return []llmguard.Finding{{
				Entity:     d.entity,
				Start:      start,
				End:        start + len(d.keyword),
				Confidence: 0.95,
			}}, nil
		}
		start++
	}
	return nil, nil
}

func TestDetector_WhenRegisteredViaWithDetector_ExpectPublicContract(t *testing.T) {
	t.Parallel()

	custom := &keywordDetector{
		name:    "keyword-email",
		keyword: "@",
		entity:  llmguard.EntityEmail,
	}

	guard, err := llmguard.New(llmguard.WithDetector(custom))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "reach me at a@b.c")
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityEmail, findings[0].Entity)
	assert.Equal(t, "keyword-email", findings[0].Detector)
}

func TestDetector_WhenMultipleCustomDetectors_ExpectAggregatedResults(t *testing.T) {
	t.Parallel()

	email := &keywordDetector{name: "email", keyword: "@", entity: llmguard.EntityEmail}
	phone := &keywordDetector{name: "phone", keyword: "555", entity: llmguard.EntityPhone}

	guard, err := llmguard.New(llmguard.WithDetector(email), llmguard.WithDetector(phone))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "call 555 or mail a@b")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(findings), 2)
}

func TestDetector_WhenNameStable_ExpectSameMetadataAcrossCalls(t *testing.T) {
	t.Parallel()

	detector := &keywordDetector{name: "stable", keyword: "x", entity: llmguard.EntityURL}
	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		findings, detectErr := guard.Detect(context.Background(), "x")
		require.NoError(t, detectErr)
		require.Len(t, findings, 1)
		assert.Equal(t, "stable", findings[0].Detector)
	}
}
