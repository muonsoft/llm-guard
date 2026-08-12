package llmguard_test

import (
	"context"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newURLGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewURLDetector()))
	require.NoError(t, err)
	return guard
}

func TestURL_WhenBracketedIPv6Host_ExpectFullURLSpan(t *testing.T) {
	t.Parallel()

	guard := newURLGuard(t)
	text := "see https://[2001:db8::1]/path?q=1"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "https://[2001:db8::1]/path?q=1", text[findings[0].Start:findings[0].End])
}

func TestURL_WhenCredentialsAndQuery_ExpectSpanWithoutTrailingPunctuation(t *testing.T) {
	t.Parallel()

	guard := newURLGuard(t)
	text := "see https://user:pass@example.com:8443/a?q=one#part)."
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)

	expected := "https://user:pass@example.com:8443/a?q=one#part"
	assert.Equal(t, expected, text[findings[0].Start:findings[0].End])
}

func TestURL_WhenUnsupportedOrMalformed_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newURLGuard(t)
	cases := []string{
		"/relative/path",
		"ftp://example.com",
		"http://localhost",
		"https://exam ple.com",
		"prefixhttps://example.comsuffix",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestURL_WhenIPHost_ExpectAccepted(t *testing.T) {
	t.Parallel()

	guard := newURLGuard(t)
	text := "ping https://192.0.2.1/status"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "https://192.0.2.1/status", text[findings[0].Start:findings[0].End])
}

func TestURL_WhenUnicodeContext_ExpectValidOffsets(t *testing.T) {
	t.Parallel()

	guard := newURLGuard(t)
	text := "ссылка https://example.com/path здесь"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, strings.Index(text, "https://example.com/path"), findings[0].Start)
}

func TestURL_WhenDetectorNameStable_ExpectURL(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "url", llmguard.NewURLDetector().Name())
}
