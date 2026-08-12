package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressDetector_WhenMandatoryForms_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAddressDetector()
	cases := []struct {
		input string
		want  string
	}{
		{"г. Москва, ул. Тверская, д. 18", "г. Москва, ул. Тверская, д. 18"},
		{"Москва, Тверская улица, дом 18", "Москва, Тверская улица, дом 18"},
		{"ул. Ленина, 10", "ул. Ленина, 10"},
		{"проспект Мира, д. 101", "проспект Мира, д. 101"},
		{"ул. Ленина, д. 15, корп. 2, стр. 1, кв. 27", "ул. Ленина, д. 15, корп. 2, стр. 1, кв. 27"},
		{"ул.Тверская,д.1", "ул.Тверская,д.1"},
		{"ул. Ленина, д. 10/2", "ул. Ленина, д. 10/2"},
		{"ул. Ленина, д. 10-2", "ул. Ленина, д. 10-2"},
		{"ул. Ленина, д. 10А", "ул. Ленина, д. 10А"},
		{"Тверская ул., 18", "Тверская ул., 18"},
		{"Мира пр-т, 18", "Мира пр-т, 18"},
		{"Садовый пер., 18", "Садовый пер., 18"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			findings, err := detector.Detect(context.Background(), tc.input)
			require.NoError(t, err)
			require.Len(t, findings, 1)
			assert.Equal(t, llmguard.EntityAddress, findings[0].Entity)
			assert.Equal(t, "address", findings[0].Detector)
			assert.Equal(t, tc.want, tc.input[findings[0].Start:findings[0].End])
		})
	}
}

func TestAddressDetector_WhenConservativeNegatives_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAddressDetector()
	for _, input := range []string{
		"Москва",
		"Санкт-Петербург",
		"Ленинградская область",
		"Тверская",
		"улица Ленина",
		"Москва, Тверская улица",
		"42",
		"xxxул. Ленина, 10",
		"ул. Ленина,,,,10",
		"ул. Ленина, д. 15,,,,корп. 2",
		"ул. Ленина, д. 10АБ",
		"ул. Ленина, д. 10/2/3",
		"ул. Ленина, д. 10-2-3",
	} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			findings, err := detector.Detect(context.Background(), input)
			require.NoError(t, err)
			assert.Empty(t, findings)
		})
	}
}

func TestAddressDetector_WhenMultipleAddresses_ExpectStableOrder(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAddressDetector()
	text := "ул. Ленина, 10 и ул. Мира, 20"
	findings, err := detector.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "ул. Ленина, 10", text[findings[0].Start:findings[0].End])
	assert.Equal(t, "ул. Мира, 20", text[findings[1].Start:findings[1].End])
}

func TestAddressDetector_WhenUnicodeEmbedded_ExpectInnerSpanOnly(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAddressDetector()
	text := "«ул. Ленина, 10» указан"
	findings, err := detector.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "ул. Ленина, 10", text[findings[0].Start:findings[0].End])
}

func TestAddressDetector_WhenContextCancelledBeforeDetect_ExpectError(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAddressDetector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	findings, err := detector.Detect(ctx, "ул. Ленина, 10")
	require.Error(t, err)
	assert.Nil(t, findings)
}

func TestAddressDetector_WhenConcurrent_ExpectDeterministicResults(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewAddressDetector()
	text := "ул. Ленина, 10 и ул. Мира, 20"

	const workers = 16
	type workerResult struct {
		findings []llmguard.Finding
		err      error
	}
	results := make(chan workerResult, workers)
	for i := 0; i < workers; i++ {
		go func() {
			findings, err := detector.Detect(context.Background(), text)
			results <- workerResult{findings: findings, err: err}
		}()
	}

	var baseline []llmguard.Finding
	for i := 0; i < workers; i++ {
		res := <-results
		require.NoError(t, res.err)
		if i == 0 {
			baseline = res.findings
			continue
		}
		assert.Equal(t, baseline, res.findings)
	}
}

func TestAddressResolve_WhenPermutedCandidates_ExpectAddressWins(t *testing.T) {
	t.Parallel()

	text := "ул. Академика Сахарова, 10"
	address := llmguard.Finding{
		Entity: llmguard.EntityAddress, Start: 0, End: 45, Confidence: 0.84, Detector: "address",
	}
	person := llmguard.Finding{
		Entity: llmguard.EntityPerson, Start: 25, End: 41, Confidence: 0.99, Detector: "person",
	}

	perms := [][]llmguard.Finding{{person, address}, {address, person}}
	var baseline []llmguard.Finding
	for i, input := range perms {
		resolved, err := llmguard.Resolve(text, input)
		require.NoError(t, err)
		require.Len(t, resolved, 1)
		assert.Equal(t, llmguard.EntityAddress, resolved[0].Entity)
		assert.Equal(t, 0, resolved[0].Start)
		assert.Equal(t, 45, resolved[0].End)
		if i == 0 {
			baseline = resolved
		} else {
			assert.Equal(t, baseline, resolved)
		}
	}
}

func TestAddressGuard_WhenSakharovStreet_ExpectNoPersonLeakage(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewAddressDetector()),
		llmguard.WithDetector(llmguard.NewPersonDetector()),
	)
	require.NoError(t, err)

	text := "ул. Академика Сахарова, 10"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, llmguard.EntityAddress, result.Findings[0].Entity)
	assert.Equal(t, 1, stringsCount(result.Text, "{{LLMG_"))
}

func TestAddressMaskRestore_WhenFullAddress_ExpectSingleTokenRoundTrip(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewAddressDetector()))
	require.NoError(t, err)

	text := "Доставка: ул. Ленина, 10."
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, 1, stringsCount(result.Text, "{{LLMG_"))

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestAddressMaskRestore_WhenConcurrentCallers_ExpectIndependentTokenSets(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewAddressDetector()))
	require.NoError(t, err)

	const workers = 8
	type workerOutcome struct {
		text        string
		maskedToken string
		tokens      *llmguard.TokenSet
		maskErr     error
		restoreErr  error
		restored    string
	}
	results := make(chan workerOutcome, workers)
	for i := 0; i < workers; i++ {
		go func(n int) {
			var outcome workerOutcome
			outcome.text = "ул. Ленина, 10"
			if n%2 == 0 {
				outcome.text = "проспект Мира, 101"
			}
			masked, err := guard.Mask(context.Background(), outcome.text)
			outcome.maskErr = err
			if err != nil {
				results <- outcome
				return
			}
			outcome.maskedToken = extractMaskedToken(masked.Text)
			outcome.tokens = masked.Tokens
			outcome.restored, outcome.restoreErr = guard.Restore(context.Background(), masked.Text, masked.Tokens)
			results <- outcome
		}(i)
	}

	tokenSets := make([]*llmguard.TokenSet, 0, workers)
	maskedTokens := make([]string, 0, workers)
	for i := 0; i < workers; i++ {
		res := <-results
		require.NoError(t, res.maskErr)
		require.NoError(t, res.restoreErr)
		assert.Equal(t, res.text, res.restored)
		require.NotNil(t, res.tokens)
		require.NotEmpty(t, res.maskedToken)
		tokenSets = append(tokenSets, res.tokens)
		maskedTokens = append(maskedTokens, res.maskedToken)
	}
	for i := 0; i < len(tokenSets); i++ {
		for j := i + 1; j < len(tokenSets); j++ {
			assert.NotSame(t, tokenSets[i], tokenSets[j])
			assert.NotEqual(t, maskedTokens[i], maskedTokens[j])
		}
	}
}
