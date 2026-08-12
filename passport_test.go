package llmguard_test

import (
	"context"
	"sync"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPassportGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewPassportDetector()))
	require.NoError(t, err)
	return guard
}

func TestPassport_WhenSupportedForms_ExpectFindings(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	text := "паспорт 45 08 123456 и паспорт РФ 4508 654321"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	for _, finding := range findings {
		assert.Equal(t, llmguard.EntityPassport, finding.Entity)
	}
}

func TestPassport_WhenSeparatedSeriesNumber_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	text := "паспорт: серия 45 08, номер 123456"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestPassport_WhenNoMarker_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "two_two_six", text: "45 08 123456"},
		{name: "four_six", text: "4508 654321"},
		{name: "contract_digits", text: "договор 4508123456"},
		{name: "embedded_digits", text: "id 45081234567"},
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

func TestPassport_WhenMalformed_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "dash_separator", text: "паспорт 4508-654321"},
		{name: "wrong_grouping", text: "паспорт 450 8654321"},
		{name: "too_short", text: "паспорт 450812345"},
		{name: "too_long", text: "паспорт 45081234567"},
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

func TestPassport_WhenMultilineContext_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	text := "паспорт\n45 08 123456"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestPassport_WhenUnicodeBoundary_ExpectValidOffsets(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	text := "данные: паспорт 45 08 123456."
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.True(t, findings[0].End <= len(text))
}

func TestPassport_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	guard := newPassportGuard(t)
	text := "паспорт 45 08 123456"

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

func TestPassport_WhenContextPreCanceled_ExpectCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := llmguard.NewPassportDetector().Detect(ctx, "паспорт 45 08 123456")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPassport_WhenDetectorNameStable_ExpectPassport(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "passport", llmguard.NewPassportDetector().Name())
}
