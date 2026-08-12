package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSNILSGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewSNILSDetector()))
	require.NoError(t, err)
	return guard
}

func TestSNILS_WhenCompactAndFormatted_ExpectBothSpans(t *testing.T) {
	t.Parallel()

	guard := newSNILSGuard(t)
	text := "СНИЛС 11223344595 и 123-456-789 64"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "11223344595", text[findings[0].Start:findings[0].End])
	assert.Equal(t, "123-456-789 64", text[findings[1].Start:findings[1].End])
}

func TestSNILS_WhenFailedChecksum_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newSNILSGuard(t)
	text := "СНИЛС 123-456-789 00"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestSNILS_WhenLegacyRange_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newSNILSGuard(t)
	text := "СНИЛС 000-000-001 00"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestSNILS_WhenNumericExtensionProbe_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newSNILSGuard(t)
	findings, err := guard.Detect(context.Background(), "123-456-789 64-1")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestSNILS_WhenMixedSeparators_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newSNILSGuard(t)
	text := "СНИЛС 123-456 789 64"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestSNILS_WhenDetectorNameStable_ExpectSNILS(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "snils", llmguard.NewSNILSDetector().Name())
}
