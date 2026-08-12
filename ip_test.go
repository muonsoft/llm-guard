package llmguard_test

import (
	"context"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIPGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewIPDetector()))
	require.NoError(t, err)
	return guard
}

func TestIP_WhenIPv4MappedIPv6_ExpectSingleFullFinding(t *testing.T) {
	t.Parallel()

	guard := newIPGuard(t)
	text := "mapped ::ffff:192.0.2.1 end"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "::ffff:192.0.2.1", text[findings[0].Start:findings[0].End])
}

func TestIP_WhenAddressPrefixExtension_ExpectNoPartialFinding(t *testing.T) {
	t.Parallel()

	guard := newIPGuard(t)
	cases := []string{
		"192.0.2.42.1",
		"10.0.0.1:8080",
		"2001:db8::1:extra",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		for _, f := range findings {
			assert.NotEqual(t, "192.0.2.42", text[f.Start:f.End], "text=%q", text)
			assert.NotEqual(t, "10.0.0.1", text[f.Start:f.End], "text=%q", text)
		}
	}
}

func TestIP_WhenValidIPv4AndIPv6_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	guard := newIPGuard(t)
	text := "hosts 192.0.2.42, 2001:db8::1 and [2001:db8::2]"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 3)

	v4 := "192.0.2.42"
	v6 := "2001:db8::1"
	v6b := "2001:db8::2"
	assert.Equal(t, strings.Index(text, v4), findings[0].Start)
	assert.Equal(t, strings.Index(text, v6), findings[1].Start)
	assert.Equal(t, strings.Index(text, v6b), findings[2].Start)
	assert.Equal(t, strings.Index(text, v6b)+len(v6b), findings[2].End)
}

func TestIP_WhenMalformedCandidates_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newIPGuard(t)
	cases := []string{
		"999.1.2.3",
		"1.2.3",
		"2001:::1",
		"fe80::1%eth0",
		"01.2.3.4",
		"host192.0.2.42suffix",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestIP_WhenUnicodePunctuation_ExpectValidOffsets(t *testing.T) {
	t.Parallel()

	guard := newIPGuard(t)
	text := "сервер 10.0.0.1 готов"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "10.0.0.1", text[findings[0].Start:findings[0].End])
}

func TestIP_WhenDetectorNameStable_ExpectIP(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "ip", llmguard.NewIPDetector().Name())
}
