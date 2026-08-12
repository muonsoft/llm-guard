package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBankCardGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewBankCardDetector()))
	require.NoError(t, err)
	return guard
}

func TestBankCard_WhenCompactAndFormatted_ExpectBothSpans(t *testing.T) {
	t.Parallel()

	guard := newBankCardGuard(t)
	text := "cards 4111111111111111 and 4111 1111 1111 1111"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "4111111111111111", text[findings[0].Start:findings[0].End])
	assert.Equal(t, "4111 1111 1111 1111", text[findings[1].Start:findings[1].End])
}

func TestBankCard_WhenFailedLuhn_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankCardGuard(t)
	text := "card 4111111111111112"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankCard_WhenHomogeneousDigits_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankCardGuard(t)
	text := "card 4444444444444444"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankCard_WhenMixedSeparators_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankCardGuard(t)
	text := "card 4111-1111 1111 1111"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankCard_WhenNumericExtensionProbe_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankCardGuard(t)
	findings, err := guard.Detect(context.Background(), "4111111111111111.1")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankCard_WhenEmbeddedInLongerNumber_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankCardGuard(t)
	text := "id 14111111111111111"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankCard_WhenDetectorNameStable_ExpectBankCard(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "bank_card", llmguard.NewBankCardDetector().Name())
}
