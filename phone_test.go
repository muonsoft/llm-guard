package llmguard_test

import (
	"context"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPhoneGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewPhoneDetector()))
	require.NoError(t, err)
	return guard
}

func TestPhone_WhenRUFormattedNumbers_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	text := "Звоните +7 (999) 123-45-67 или 8 999 123 45 67"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)

	first := "+7 (999) 123-45-67"
	second := "8 999 123 45 67"
	assert.Equal(t, strings.Index(text, first), findings[0].Start)
	assert.Equal(t, strings.Index(text, first)+len(first), findings[0].End)
	assert.Equal(t, strings.Index(text, second), findings[1].Start)
	assert.Equal(t, strings.Index(text, second)+len(second), findings[1].End)
}

func TestPhone_WhenCompactInternational_ExpectAccepted(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	text := "call +442079460958 today"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "+442079460958", text[findings[0].Start:findings[0].End])
}

func TestPhone_WhenVerificationProbes_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	cases := []string{
		"+7 (9-9-9) 123-45-67",
		"+44)20(79460958",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestPhone_WhenMalformedParenthesesOrDots_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	cases := []string{
		"+7 999) 123-45-67",
		"+7 (99) 123-45-67",
		"+7 (9999) 123-45-67",
		"+7.999.123.45.67",
		"+7(999)123-45-67((extra))",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestPhone_WhenInternationalWithSeparators_ExpectAccepted(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	text := "call +44 20 7946 0958 today"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	expected := "+44 20 7946 0958"
	assert.Equal(t, expected, text[findings[0].Start:findings[0].End])
}

func TestPhone_WhenAmbiguousLongNumbers_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	cases := []string{
		"order 123456789012345",
		"9991234567",
		"+1234",
		"phone abc+79991234567def",
		"+7(999)123-45-67((extra))",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestPhone_WhenUnicodeByteOffsets_ExpectValidBoundaries(t *testing.T) {
	t.Parallel()

	guard := newPhoneGuard(t)
	text := "тел +7 999 123 45 67."
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.True(t, utf8BoundaryOK(text, findings[0].Start))
	assert.True(t, utf8BoundaryOK(text, findings[0].End))
}

func TestPhone_WhenDetectorNameStable_ExpectPhone(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "phone", llmguard.NewPhoneDetector().Name())
}
