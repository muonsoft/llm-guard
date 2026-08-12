package llmguard_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStructuredGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewPhoneDetector()),
		llmguard.WithDetector(llmguard.NewIPDetector()),
		llmguard.WithDetector(llmguard.NewURLDetector()),
		llmguard.WithDetector(llmguard.NewINNDetector()),
		llmguard.WithDetector(llmguard.NewSNILSDetector()),
		llmguard.WithDetector(llmguard.NewBankCardDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
		llmguard.WithDetector(llmguard.NewPassportDetector()),
		llmguard.WithDetector(llmguard.NewBankAccountDetector()),
		llmguard.WithDetector(llmguard.NewDateOfBirthDetector()),
	)
	require.NoError(t, err)
	return guard
}

func TestMixedStructured_WhenUnicodeText_ExpectMaskRestoreRoundTrip(t *testing.T) {
	t.Parallel()

	guard := newStructuredGuard(t)
	text := strings.Join([]string{
		"Контакт +7 (999) 123-45-67,",
		"сервер 192.0.2.42,",
		"ссылка https://example.com/path,",
		"ИНН 7707083893,",
		"СНИЛС 123-456-789 64,",
		"карта 4111111111111111,",
		"паспорт 45 08 123456,",
		"расчётный счёт 40702810900000000001,",
		"дата рождения 12.10.1990.",
	}, " ")

	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 9)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestStructured_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	guard := newStructuredGuard(t)
	text := "tel +7 999 123 45 67 ip 10.0.0.1 url https://example.com inn 7707083893 snils 123-456-789 64 card 4111111111111111 passport 45 08 123456 account 40702810900000000001 dob 12.10.1990"

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

func TestStructured_WhenFindingsFormatted_ExpectNoRawValues(t *testing.T) {
	t.Parallel()

	guard := newStructuredGuard(t)
	sensitive := "+7 (999) 123-45-67 192.0.2.42 https://secret.example.com/p 7707083893 123-456-789 64 4111111111111111"
	findings, err := guard.Detect(context.Background(), sensitive)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	for _, finding := range findings {
		payload, marshalErr := json.Marshal(finding)
		require.NoError(t, marshalErr)
		body := string(payload)
		assert.NotContains(t, body, "+7")
		assert.NotContains(t, body, "192.0.2")
		assert.NotContains(t, body, "secret.example")
		assert.NotContains(t, body, "7707083893")
		assert.NotContains(t, body, "4111111111111111")
	}
}

func TestStructured_WhenMaskResultJSON_ExpectNoRawValues(t *testing.T) {
	t.Parallel()

	guard := newStructuredGuard(t)
	text := "card 4111111111111111"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)

	payload, err := json.Marshal(result)
	require.NoError(t, err)
	body := string(payload)
	assert.NotContains(t, body, "4111111111111111")
}

type corpusCategory string

const (
	corpusPositive  corpusCategory = "positive"
	corpusNegative  corpusCategory = "negative"
	corpusMalformed corpusCategory = "malformed"
)

type corpusCase struct {
	entity   llmguard.EntityType
	text     string
	value    string
	category corpusCategory
}

type corpusStats struct {
	expectedPositive  int
	expectedNegative  int
	expectedMalformed int
	detectedPositive  int
	detectedNegative  int
	detectedMalformed int
	falsePos          int
	falseNeg          int
}

func TestStructuredCorpusEvaluation_WhenPerEntityCases_ExpectCountsMatch(t *testing.T) {
	t.Parallel()

	cases := []corpusCase{
		{entity: llmguard.EntityPhone, text: "tel +7 (999) 123-45-67", value: "+7 (999) 123-45-67", category: corpusPositive},
		{entity: llmguard.EntityPhone, text: "call +442079460958", value: "+442079460958", category: corpusPositive},
		{entity: llmguard.EntityPhone, text: "digits 9991234567", category: corpusNegative},
		{entity: llmguard.EntityPhone, text: "+7.999.123.45.67", category: corpusMalformed},
		{entity: llmguard.EntityPhone, text: "+7 (99) 123-45-67", category: corpusMalformed},

		{entity: llmguard.EntityPhone, text: "+7 (9-9-9) 123-45-67", category: corpusMalformed},
		{entity: llmguard.EntityPhone, text: "+44)20(79460958", category: corpusMalformed},

		{entity: llmguard.EntityINN, text: "7707083893-1", category: corpusMalformed},
		{entity: llmguard.EntitySNILS, text: "123-456-789 64-1", category: corpusMalformed},
		{entity: llmguard.EntityBankCard, text: "4111111111111111.1", category: corpusMalformed},

		{entity: llmguard.EntityIPAddress, text: "ip 192.0.2.42", value: "192.0.2.42", category: corpusPositive},
		{entity: llmguard.EntityIPAddress, text: "mapped ::ffff:192.0.2.1", value: "::ffff:192.0.2.1", category: corpusPositive},
		{entity: llmguard.EntityIPAddress, text: "bad 999.1.2.3", category: corpusNegative},
		{entity: llmguard.EntityIPAddress, text: "partial 192.0.2.42.1", category: corpusMalformed},

		{entity: llmguard.EntityURL, text: "u https://example.com/a", value: "https://example.com/a", category: corpusPositive},
		{entity: llmguard.EntityURL, text: "v6 https://[2001:db8::1]/p", value: "https://[2001:db8::1]/p", category: corpusPositive},
		{entity: llmguard.EntityURL, text: "rel /only/path", category: corpusNegative},
		{entity: llmguard.EntityURL, text: "bad https://exam ple.com", category: corpusMalformed},

		{entity: llmguard.EntityINN, text: "inn 7707083893", value: "7707083893", category: corpusPositive},
		{entity: llmguard.EntityINN, text: "inn 500100732259", value: "500100732259", category: corpusPositive},
		{entity: llmguard.EntityINN, text: "inn 7707083890", category: corpusNegative},
		{entity: llmguard.EntityINN, text: "inn 500100732250", category: corpusNegative},
		{entity: llmguard.EntityINN, text: "inn 1111111111", category: corpusMalformed},
		{entity: llmguard.EntityINN, text: "inn 12345678901", category: corpusMalformed},

		{entity: llmguard.EntitySNILS, text: "s 123-456-789 64", value: "123-456-789 64", category: corpusPositive},
		{entity: llmguard.EntitySNILS, text: "s 11223344595", value: "11223344595", category: corpusPositive},
		{entity: llmguard.EntitySNILS, text: "s 123-456-789 00", category: corpusNegative},
		{entity: llmguard.EntitySNILS, text: "s 123-456 789 64", category: corpusMalformed},

		{entity: llmguard.EntityBankCard, text: "c 4111111111111111", value: "4111111111111111", category: corpusPositive},
		{entity: llmguard.EntityBankCard, text: "c 4111 1111 1111 1111", value: "4111 1111 1111 1111", category: corpusPositive},
		{entity: llmguard.EntityBankCard, text: "c 4111111111111112", category: corpusNegative},
		{entity: llmguard.EntityBankCard, text: "c 4111-1111 1111 1111", category: corpusMalformed},

		{entity: llmguard.EntityPassport, text: "паспорт 45 08 123456", value: "45 08 123456", category: corpusPositive},
		{entity: llmguard.EntityPassport, text: "паспорт РФ 4508 654321", value: "4508 654321", category: corpusPositive},
		{entity: llmguard.EntityPassport, text: "4508 654321", category: corpusNegative},
		{entity: llmguard.EntityPassport, text: "паспорт: серия 45 08, номер 123456", category: corpusNegative},
		{entity: llmguard.EntityPassport, text: "паспорт 4508-654321", category: corpusMalformed},

		{entity: llmguard.EntityBankAccount, text: "расчётный счёт 40702810900000000001", value: "40702810900000000001", category: corpusPositive},
		{entity: llmguard.EntityBankAccount, text: "р/с 4070 2810 9000 0000 0001", value: "4070 2810 9000 0000 0001", category: corpusPositive},
		{entity: llmguard.EntityBankAccount, text: "40702810900000000001", category: corpusNegative},
		{entity: llmguard.EntityBankAccount, text: "договор 40702810900000000001", category: corpusNegative},
		{entity: llmguard.EntityBankAccount, text: "расчётный счёт 11111111111111111111", category: corpusMalformed},

		{entity: llmguard.EntityDateOfBirth, text: "дата рождения: 12.10.1990", value: "12.10.1990", category: corpusPositive},
		{entity: llmguard.EntityDateOfBirth, text: "родился 3 марта 1985 года", value: "3 марта 1985", category: corpusPositive},
		{entity: llmguard.EntityDateOfBirth, text: "встреча 12.10.2026", category: corpusNegative},
		{entity: llmguard.EntityDateOfBirth, text: "договор от 12.10.2026", category: corpusNegative},
		{entity: llmguard.EntityDateOfBirth, text: "дата рождения 31.02.1990", category: corpusMalformed},
	}

	allEntities := []llmguard.EntityType{
		llmguard.EntityPhone,
		llmguard.EntityIPAddress,
		llmguard.EntityURL,
		llmguard.EntityINN,
		llmguard.EntitySNILS,
		llmguard.EntityBankCard,
		llmguard.EntityPassport,
		llmguard.EntityBankAccount,
		llmguard.EntityDateOfBirth,
	}
	stats := make(map[llmguard.EntityType]*corpusStats, len(allEntities))
	for _, entity := range allEntities {
		stats[entity] = &corpusStats{}
	}
	for _, tc := range cases {
		switch tc.category {
		case corpusPositive:
			stats[tc.entity].expectedPositive++
		case corpusNegative:
			stats[tc.entity].expectedNegative++
		case corpusMalformed:
			stats[tc.entity].expectedMalformed++
		}
	}

	guard := newStructuredGuard(t)
	for _, tc := range cases {
		findings, err := guard.Detect(context.Background(), tc.text)
		require.NoError(t, err)

		var entityFindings []llmguard.Finding
		for _, finding := range findings {
			if finding.Entity == tc.entity {
				entityFindings = append(entityFindings, finding)
			}
		}

		st := stats[tc.entity]
		switch tc.category {
		case corpusPositive:
			matched := false
			for _, finding := range entityFindings {
				if tc.value == "" || tc.text[finding.Start:finding.End] == tc.value {
					matched = true
					break
				}
			}
			if matched {
				st.detectedPositive++
			} else {
				st.falseNeg++
			}
		case corpusNegative, corpusMalformed:
			if len(entityFindings) > 0 {
				st.falsePos++
			} else {
				if tc.category == corpusNegative {
					st.detectedNegative++
				} else {
					st.detectedMalformed++
				}
			}
		}
	}

	for _, entity := range allEntities {
		st := stats[entity]
		t.Logf("entity=%s expected_pos=%d detected_pos=%d expected_neg=%d detected_neg=%d expected_mal=%d detected_mal=%d false_pos=%d false_neg=%d",
			entity,
			st.expectedPositive, st.detectedPositive,
			st.expectedNegative, st.detectedNegative,
			st.expectedMalformed, st.detectedMalformed,
			st.falsePos, st.falseNeg,
		)
		assert.Equal(t, st.expectedPositive, st.detectedPositive, "entity=%s positive detected", entity)
		assert.Equal(t, st.expectedNegative, st.detectedNegative, "entity=%s negative detected", entity)
		assert.Equal(t, st.expectedMalformed, st.detectedMalformed, "entity=%s malformed detected", entity)
		assert.Zero(t, st.falsePos, "entity=%s false positives", entity)
		assert.Zero(t, st.falseNeg, "entity=%s false negatives", entity)
	}
}

func TestStructured_WhenResolveErrors_ExpectNoSensitiveSubstring(t *testing.T) {
	t.Parallel()

	sensitive := "4111111111111111"
	err := func() error {
		_, resolveErr := llmguard.Resolve(sensitive, []llmguard.Finding{{
			Entity: llmguard.EntityBankCard, Start: 0, End: 5, Confidence: 2, Detector: "bank_card",
		}})
		return resolveErr
	}()
	require.Error(t, err)
	assert.NotContains(t, err.Error(), sensitive)

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "confidence", inv.Field)
}

func TestResolve_WhenURLOverlapsEmail_ExpectURLWins(t *testing.T) {
	t.Parallel()

	urlText := "see https://user:pass@mail.example.com/path?q=a@b.co"
	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewURLDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
	)
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), urlText)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	resolved, err := llmguard.Resolve(urlText, findings)
	require.NoError(t, err)

	var hasURL, hasEmail bool
	for _, finding := range resolved {
		switch finding.Entity {
		case llmguard.EntityURL:
			hasURL = true
		case llmguard.EntityEmail:
			hasEmail = true
		}
	}
	assert.True(t, hasURL)
	assert.False(t, hasEmail)
}

func TestResolve_WhenStructuredNumericOverlapsCustom_ExpectStructuredWins(t *testing.T) {
	t.Parallel()

	payload := "card 4111111111111111"
	custom := llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM_NUM"), Start: 5, End: 21, Confidence: 1, Detector: "custom",
	}
	bank := llmguard.Finding{
		Entity: llmguard.EntityBankCard, Start: 5, End: 21, Confidence: 0.87, Detector: "bank_card",
	}

	perms := [][]llmguard.Finding{{custom, bank}, {bank, custom}}
	for _, input := range perms {
		resolved, err := llmguard.Resolve(payload, input)
		require.NoError(t, err)
		require.Len(t, resolved, 1)
		assert.Equal(t, llmguard.EntityBankCard, resolved[0].Entity)
	}
}

func TestResolve_WhenINNOverlapsCustom_ExpectINNWins(t *testing.T) {
	t.Parallel()

	payload := "id 7707083893"
	inn := llmguard.Finding{
		Entity: llmguard.EntityINN, Start: 3, End: 13, Confidence: 0.92, Detector: "inn",
	}
	custom := llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM_NUM"), Start: 3, End: 13, Confidence: 1, Detector: "custom",
	}
	resolved, err := llmguard.Resolve(payload, []llmguard.Finding{custom, inn})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntityINN, resolved[0].Entity)
}

func TestResolve_WhenSNILSOverlapsCustom_ExpectSNILSWins(t *testing.T) {
	t.Parallel()

	payload := "sn 123-456-789 64"
	snils := llmguard.Finding{
		Entity: llmguard.EntitySNILS, Start: 3, End: 17, Confidence: 0.91, Detector: "snils",
	}
	custom := llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM_NUM"), Start: 3, End: 17, Confidence: 1, Detector: "custom",
	}
	resolved, err := llmguard.Resolve(payload, []llmguard.Finding{custom, snils})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntitySNILS, resolved[0].Entity)
}

func TestResolve_WhenURLEmailPermuted_ExpectStableURLWinner(t *testing.T) {
	t.Parallel()

	text := "x https://a@b.co/y z"
	urlFinding := llmguard.Finding{
		Entity: llmguard.EntityURL, Start: 2, End: 18, Confidence: 0.88, Detector: "url",
	}
	emailFinding := llmguard.Finding{
		Entity: llmguard.EntityEmail, Start: 10, End: 16, Confidence: 0.9, Detector: "email",
	}

	first, err := llmguard.Resolve(text, []llmguard.Finding{urlFinding, emailFinding})
	require.NoError(t, err)
	second, err := llmguard.Resolve(text, []llmguard.Finding{emailFinding, urlFinding})
	require.NoError(t, err)
	assert.Equal(t, first, second)
	require.Len(t, first, 1)
	assert.Equal(t, llmguard.EntityURL, first[0].Entity)
}
