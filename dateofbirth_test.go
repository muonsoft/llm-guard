package llmguard_test

import (
	"context"
	"sync"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDateOfBirthGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewDateOfBirthDetector()))
	require.NoError(t, err)
	return guard
}

func TestDateOfBirth_WhenNumericAndTextual_ExpectFindings(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "дата рождения: 12.10.1990 и родился 3 марта 1985 года"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, llmguard.EntityDateOfBirth, findings[0].Entity)
	assert.Equal(t, llmguard.EntityDateOfBirth, findings[1].Entity)
}

func TestDateOfBirth_WhenSlashForm_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "д.р. 01/05/2000"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityDateOfBirth, findings[0].Entity)
}

func TestDateOfBirth_WhenOrdinaryDates_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "meeting_date", text: "встреча 12.10.2026"},
		{name: "contract_date", text: "договор от 12.10.2026"},
		{name: "deadline_textual", text: "срок 3 марта 2026 года"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings, err := guard.Detect(context.Background(), tc.text)
			require.NoError(t, err)
			assert.Empty(t, findings)
		})
	}
}

func TestDateOfBirth_WhenImpossibleDate_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "invalid_february_day", text: "дата рождения 31.02.1990"},
		{name: "invalid_day", text: "родился 32 января 1990"},
		{name: "year_extension", text: "родилась 1 февраля 19900"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings, err := guard.Detect(context.Background(), tc.text)
			require.NoError(t, err)
			assert.Empty(t, findings)
		})
	}
}

func TestDateOfBirth_WhenMultilineContext_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "дата рождения\n12.10.1990"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestDateOfBirth_WhenMultilineTextualCandidate_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "родился 3 марта\n1985 года"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestDateOfBirth_WhenEmbeddedDate_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "leading_letter", text: "дата рождения x12.10.1990"},
		{name: "trailing_letter", text: "дата рождения 12.10.1990y"},
		{name: "extra_numeric_component", text: "дата рождения 12.10.1990.2020"},
		{name: "chained_numeric_component", text: "дата рождения 01.12.10.1990"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			findings, err := guard.Detect(context.Background(), tc.text)
			require.NoError(t, err)
			assert.Empty(t, findings)
		})
	}
}

func TestDateOfBirth_WhenFemaleMarker_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "родилась 15 апреля 1992"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityDateOfBirth, findings[0].Entity)
}

func TestDateOfBirth_WhenUnicodeBoundary_ExpectValidOffsets(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "поле: дата рождения 12.10.1990."
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.True(t, findings[0].End <= len(text))
}

func TestDateOfBirth_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	guard := newDateOfBirthGuard(t)
	text := "дата рождения 12.10.1990"

	const workers = 16
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := guard.Detect(context.Background(), text)
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestDateOfBirth_WhenContextPreCanceled_ExpectCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := llmguard.NewDateOfBirthDetector().Detect(ctx, "дата рождения 12.10.1990")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDateOfBirth_WhenDetectorNameStable_ExpectDateOfBirth(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "date_of_birth", llmguard.NewDateOfBirthDetector().Name())
}
