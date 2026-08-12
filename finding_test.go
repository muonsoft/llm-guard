package llmguard_test

import (
	"context"
	"math"
	"testing"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type findingDetector struct {
	name     string
	findings []llmguard.Finding
}

func (d *findingDetector) Name() string { return d.name }

func (d *findingDetector) Detect(context.Context, string) ([]llmguard.Finding, error) {
	return d.findings, nil
}

func detectFinding(t *testing.T, text string, finding llmguard.Finding) error {
	t.Helper()

	guard, err := llmguard.New(llmguard.WithDetector(&findingDetector{
		name:     "test",
		findings: []llmguard.Finding{finding},
	}))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), text)
	return err
}

func TestFinding_WhenUnicodeSpanValid_ExpectAccepted(t *testing.T) {
	t.Parallel()

	text := "café"
	start := len("caf")
	end := len(text)

	err := detectFinding(t, text, llmguard.Finding{
		Entity: llmguard.EntityPerson, Start: start, End: end, Confidence: 1,
	})
	require.NoError(t, err)
}

func TestFinding_WhenSpanInsideRune_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	text := "café"
	err := detectFinding(t, text, llmguard.Finding{
		Entity: llmguard.EntityPerson, Start: 2, End: 4, Confidence: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "utf8_boundary", inv.Field)
	assert.NotContains(t, err.Error(), "caf")
}

func TestFinding_WhenEmptyEntity_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: "", Start: 0, End: 1, Confidence: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "entity", inv.Field)
}

func TestFinding_WhenEmptySpan_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 0, Confidence: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "span", inv.Field)
}

func TestFinding_WhenSpanOutOfRange_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "ab", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 5, Confidence: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "span", inv.Field)
}

func TestFinding_WhenNegativeStart_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: -1, End: 1, Confidence: 1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
}

func TestFinding_WhenNaNConfidence_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: math.NaN(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "confidence", inv.Field)
}

func TestFinding_WhenInfinityConfidence_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: math.Inf(1),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
}

func TestFinding_WhenConfidenceBelowZero_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: -0.1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
}

func TestFinding_WhenConfidenceAboveOne_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1.1,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
}

func TestFinding_WhenEndAtTextLength_ExpectAccepted(t *testing.T) {
	t.Parallel()

	text := "hello"
	err := detectFinding(t, text, llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: len(text), Confidence: 1,
	})
	require.NoError(t, err)
}

func TestFinding_WhenCustomEntity_ExpectAccepted(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "x", llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM"), Start: 0, End: 1, Confidence: 0.5,
	})
	require.NoError(t, err)
}

func TestFinding_WhenInvalidMetadata_ExpectInvalidFinding(t *testing.T) {
	t.Parallel()

	err := detectFinding(t, "a", llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 0, End: 1, Confidence: 1, Detector: "wrong",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "detector", inv.Field)
}

func TestFinding_WhenErrorMessage_ExpectNoInputSubstring(t *testing.T) {
	t.Parallel()

	sensitive := "super-secret-token"
	err := detectFinding(t, sensitive, llmguard.Finding{
		Entity: llmguard.EntitySecretAPIKey, Start: 0, End: 5, Confidence: 2,
	})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), sensitive)
	assert.NotContains(t, err.Error(), sensitive[:5])
}
