package llmguard_test

import (
	"context"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonDetector_WhenMandatoryForms_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewPersonDetector()
	cases := []struct {
		input string
		want  string
	}{
		{"Иван Петров", "Иван Петров"},
		{"Петров Иван", "Петров Иван"},
		{"Иван Сергеевич Петров", "Иван Сергеевич Петров"},
		{"Петров Иван Сергеевич", "Петров Иван Сергеевич"},
		{"Петров И. С.", "Петров И. С."},
		{"И. С. Петров", "И. С. Петров"},
		{"Ивану Сергеевичу Петрову", "Ивану Сергеевичу Петрову"},
		{"Иваном Сергеевичем Петровым", "Иваном Сергеевичем Петровым"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			findings, err := detector.Detect(context.Background(), tc.input)
			require.NoError(t, err)
			require.Len(t, findings, 1)
			assert.Equal(t, llmguard.EntityPerson, findings[0].Entity)
			assert.Equal(t, "person", findings[0].Detector)
			assert.Equal(t, tc.want, tc.input[findings[0].Start:findings[0].End])
		})
	}
}

func TestPersonDetector_WhenConservativeNegatives_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewPersonDetector()
	for _, input := range []string{
		"Иван",
		"Петров",
		"директор Иван",
		"улица Гагарина",
		"иван петров",
		"xxxИван Петров",
		"Проект Альфа",
		"Большой Театр",
		"пр-т Гагарина",
		"ул. Петрова",
		"Иван Новый",
		"Сергей Синий",
		"Мария Новая",
		"Петров И . С .",
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

func TestPersonDetector_WhenMultiplePersons_ExpectStableOrder(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewPersonDetector()
	text := "Иван Петров встретил Петров Иван"
	findings, err := detector.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "Иван Петров", text[findings[0].Start:findings[0].End])
	assert.Equal(t, "Петров Иван", text[findings[1].Start:findings[1].End])
}

func TestPersonDetector_WhenContextCancelledBeforeDetect_ExpectError(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewPersonDetector()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	findings, err := detector.Detect(ctx, "Иван Петров")
	require.Error(t, err)
	assert.Nil(t, findings)
}

func TestPersonDetector_WhenConcurrent_ExpectDeterministicResults(t *testing.T) {
	t.Parallel()

	detector := llmguard.NewPersonDetector()
	text := "Иван Петров встретил Петров Иван"

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

func TestPersonResolve_WhenOverlapWithLowerPriority_ExpectPersonWins(t *testing.T) {
	t.Parallel()

	text := "Иван Петров"
	person := llmguard.Finding{
		Entity: llmguard.EntityPerson, Start: 0, End: 21, Confidence: 0.82, Detector: "person",
	}
	phone := llmguard.Finding{
		Entity: llmguard.EntityPhone, Start: 9, End: 21, Confidence: 1, Detector: "phone",
	}

	resolved, err := llmguard.Resolve(text, []llmguard.Finding{phone, person})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntityPerson, resolved[0].Entity)
	assert.Equal(t, 0, resolved[0].Start)
	assert.Equal(t, 21, resolved[0].End)
}

func TestPersonMaskRestore_WhenFullName_ExpectSingleTokenRoundTrip(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(llmguard.NewPersonDetector()))
	require.NoError(t, err)

	text := "Документ подписал Иван Петров."
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, 1, stringsCount(result.Text, "{{LLMG_"))

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestPersonMaskRestore_WhenOverlapWithEmail_ExpectBothMasked(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(
		llmguard.WithDetector(llmguard.NewPersonDetector()),
		llmguard.WithDetector(llmguard.NewEmailDetector()),
	)
	require.NoError(t, err)

	text := "Иван Петров a@b.co"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 2)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func stringsCount(text, needle string) int {
	count := 0
	for i := 0; i <= len(text)-len(needle); i++ {
		if text[i:i+len(needle)] == needle {
			count++
		}
	}
	return count
}
