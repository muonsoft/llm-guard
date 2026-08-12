package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newINNGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewINNDetector()))
	require.NoError(t, err)
	return guard
}

func TestINN_WhenValidLegalAndIndividual_ExpectTwoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	text := "ИНН 7707083893 и 500100732259"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, llmguard.EntityINN, findings[0].Entity)
	assert.Equal(t, "7707083893", text[findings[0].Start:findings[0].End])
	assert.Equal(t, "500100732259", text[findings[1].Start:findings[1].End])
}

func TestINN_WhenFailedChecksum12Digit_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	text := "ИНН 500100732250"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestINN_WhenHomogeneousDigits_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	cases := []string{
		"1111111111",
		"222222222222",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestINN_WhenFailedChecksum_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	text := "ИНН 7707083890"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestINN_WhenNumericExtensionProbe_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	findings, err := guard.Detect(context.Background(), "7707083893-1")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestINN_WhenEmbeddedInLongerNumber_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	text := "id 177070838931"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestINN_WhenMalformedLength_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newINNGuard(t)
	cases := []string{"123", "12345678901", "1234567890123"}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestINN_WhenDetectorNameStable_ExpectINN(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "inn", llmguard.NewINNDetector().Name())
}
