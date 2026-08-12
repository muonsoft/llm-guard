package llmguard_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEmailGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewEmailDetector()))
	require.NoError(t, err)
	return guard
}

func TestEmail_WhenCommonMailboxWithPunctuation_ExpectExactSpan(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	text := "Напишите (Ivan.Petrov+sales@example-domain.ru)."
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)

	expected := "Ivan.Petrov+sales@example-domain.ru"
	start := strings.Index(text, expected)
	require.NotEqual(t, -1, start)

	assert.Equal(t, llmguard.EntityEmail, findings[0].Entity)
	assert.Equal(t, start, findings[0].Start)
	assert.Equal(t, start+len(expected), findings[0].End)
	assert.Equal(t, "email", findings[0].Detector)
	assert.InDelta(t, 0.9, findings[0].Confidence, 0)
}

func TestEmail_WhenMultipleMailboxesInUnicodeText_ExpectCorrectOffsets(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	text := "Контакты: a@b.co и z.y@x.org тут"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)

	first := "a@b.co"
	second := "z.y@x.org"
	assert.Equal(t, strings.Index(text, first), findings[0].Start)
	assert.Equal(t, strings.Index(text, first)+len(first), findings[0].End)
	assert.Equal(t, strings.Index(text, second), findings[1].Start)
	assert.Equal(t, strings.Index(text, second)+len(second), findings[1].End)
}

func TestEmail_WhenInvalidForms_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	cases := []string{
		".user@example.com",
		"user..name@example.com",
		"user@-example.com",
		"user@example",
		"Жuser@example.com",
		"user@example.comЖ",
		"user@example.com_",
		"just an @ sign",
		"http://example.com/path",
		"{{PII_EMAIL_0001}}",
	}
	for _, text := range cases {
		findings, err := guard.Detect(context.Background(), text)
		require.NoError(t, err, "text=%q", text)
		assert.Empty(t, findings, "text=%q", text)
	}
}

func TestEmail_WhenDNSLikeSuffixes_ExpectAccepted(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	cases := []struct {
		text    string
		mailbox string
	}{
		{"user@example.company", "user@example.company"},
		{"buy@brand.coffee", "buy@brand.coffee"},
		{"team@mail.community", "team@mail.community"},
	}
	for _, tc := range cases {
		findings, err := guard.Detect(context.Background(), tc.text)
		require.NoError(t, err, "text=%q", tc.text)
		require.Len(t, findings, 1, "text=%q", tc.text)
		assert.Equal(t, strings.Index(tc.text, tc.mailbox), findings[0].Start)
		assert.Equal(t, strings.Index(tc.text, tc.mailbox)+len(tc.mailbox), findings[0].End)
	}
}

func TestEmail_WhenPreservesCase_ExpectOriginalCaseSpan(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	text := "Write to Foo.Bar@Example.COM now"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "Foo.Bar@Example.COM", text[findings[0].Start:findings[0].End])
}

func TestEmail_WhenContextCanceled_ExpectNoPartialFindings(t *testing.T) {
	t.Parallel()

	blocking := &blockingEmailDetector{started: make(chan struct{})}
	guard, err := llmguard.New(llmguard.WithDetector(blocking))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var detectErr error
	go func() {
		_, detectErr = guard.Detect(ctx, "a@b.co")
		close(done)
	}()

	<-blocking.started
	cancel()
	<-done
	require.Error(t, detectErr)
	assert.ErrorIs(t, detectErr, context.Canceled)
}

type blockingEmailDetector struct {
	started chan struct{}
}

func (blockingEmailDetector) Name() string { return "blocking-email" }

func (d *blockingEmailDetector) Detect(ctx context.Context, text string) ([]llmguard.Finding, error) {
	close(d.started)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestEmail_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	const workers = 16
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			_, err := guard.Detect(context.Background(), "ping a@b.co and c@d.org")
			errCh <- err
		}()
	}
	for i := 0; i < workers; i++ {
		require.NoError(t, <-errCh)
	}
}

func TestEmail_WhenDetectorNameStable_ExpectEmail(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewEmailDetector()
	assert.Equal(t, "email", detector.Name())
}

func TestEmail_WhenValidUTF8Boundaries_ExpectAcceptedByCore(t *testing.T) {
	t.Parallel()

	guard := newEmailGuard(t)
	text := "café a@b.co done"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.True(t, utf8BoundaryOK(text, findings[0].Start))
	assert.True(t, utf8BoundaryOK(text, findings[0].End))
}

func utf8BoundaryOK(text string, index int) bool {
	if index < 0 || index > len(text) {
		return false
	}
	if index == len(text) {
		return true
	}
	return text[index]&0xC0 != 0x80
}

func TestEmail_WhenContextDeadlineExceeded_ExpectDeadlineExceeded(t *testing.T) {
	t.Parallel()

	blocking := &blockingEmailDetector{started: make(chan struct{})}
	guard, err := llmguard.New(llmguard.WithDetector(blocking))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	var detectErr error
	go func() {
		_, detectErr = guard.Detect(ctx, "a@b.co")
		close(done)
	}()

	<-blocking.started
	<-done
	require.Error(t, detectErr)
	assert.ErrorIs(t, detectErr, context.DeadlineExceeded)
}
