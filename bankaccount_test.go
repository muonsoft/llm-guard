package llmguard_test

import (
	"context"
	"sync"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validBIK                 = "044525225"
	validBankAccount         = "40702810000000000007"
	invalidBankAccount       = "40702810900000000001"
	groupedBankAccount       = "4070 2810 0000 0000 0007"
	groupedBankAccountDigits = "40702810000000000007"
)

func newBankAccountGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewBankAccountDetector()))
	require.NoError(t, err)
	return guard
}

func TestBankAccount_WhenContextConfirmedWithoutBIK_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "расчётный счёт " + invalidBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenBIKBeforeAccountValidChecksum_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "БИК " + validBIK + " расчётный счёт " + validBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenBIKAfterAccountValidChecksum_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "расчётный счёт " + validBankAccount + " БИК " + validBIK
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenBIKInvalidChecksumBeforeAccount_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "БИК " + validBIK + " расчётный счёт " + invalidBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankAccount_WhenBIKInvalidChecksumAfterAccount_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "расчётный счёт " + invalidBankAccount + " БИК " + validBIK
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestBankAccount_WhenFalseBIKSubstring_ExpectContextFallback(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "словоабик " + validBIK + " расчётный счёт " + invalidBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenFalseBIKLabelWithWordGap_ExpectContextFallback(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "БИК банка " + validBIK + " расчётный счёт " + invalidBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenUnlabeledNineDigitsOnLine_ExpectContextFallback(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "код 123456789 расчётный счёт " + invalidBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenGroupedForm_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "р/с " + groupedBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
}

func TestBankAccount_WhenContractNumber_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "contract_without_marker", text: "договор " + invalidBankAccount},
		{name: "bare_account", text: invalidBankAccount},
		{name: "homogeneous_digits", text: "банковский счёт 11111111111111111111"},
		{name: "mixed_separators", text: "расчётный счёт 4070-2810-9000-0000-0001"},
		{name: "numeric_extension", text: "расчётный счёт 407028109000000000011"},
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

func TestBankAccount_WhenMalformed_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	cases := []struct {
		name string
		text string
	}{
		{name: "too_short", text: "расчётный счёт 123"},
		{name: "incomplete_groups", text: "расчётный счёт 4070 2810 9000 0001"},
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

func TestBankAccount_WhenUnicodeBoundary_ExpectValidOffsets(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "реквизиты: расчётный счёт " + invalidBankAccount + "."
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.True(t, findings[0].End <= len(text))
}

func TestBankAccount_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "расчётный счёт " + invalidBankAccount

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

func TestBankAccount_WhenContextPreCanceled_ExpectCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := llmguard.NewBankAccountDetector().Detect(ctx, "расчётный счёт 40702810900000000001")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBankAccount_WhenDetectorNameStable_ExpectBankAccount(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "bank_account", llmguard.NewBankAccountDetector().Name())
}

func TestBankAccount_WhenGroupedValidWithBIK_ExpectFinding(t *testing.T) {
	t.Parallel()

	guard := newBankAccountGuard(t)
	text := "БИК " + validBIK + " банковский счёт " + groupedBankAccount
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, llmguard.EntityBankAccount, findings[0].Entity)
	assert.Equal(t, groupedBankAccountDigits, collectDigits(text[findings[0].Start:findings[0].End]))
}

func collectDigits(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b = append(b, s[i])
		}
	}
	return string(b)
}
