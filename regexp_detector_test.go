package llmguard_test

import (
	"context"
	"math"
	"strings"
	"sync"
	"testing"

	muerrors "github.com/muonsoft/errors"
	"github.com/muonsoft/llm-guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegex_WhenValidConfig_ExpectDetector(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-[0-9]{6}`,
		Confidence: 0.9,
	})
	require.NoError(t, err)
	assert.Equal(t, "employee_id", detector.Name())
}

func TestRegex_WhenInvalidPattern_ExpectSafeConfigError(t *testing.T) {
	t.Parallel()

	sensitivePattern := `EMP-[`
	_, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    sensitivePattern,
		Confidence: 0.9,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
	assert.NotContains(t, err.Error(), sensitivePattern)
	assert.NotContains(t, err.Error(), "EMPLOYEE_ID")
}

func TestRegex_WhenEmptyFields_ExpectSafeConfigError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		config llmguard.RegexDetectorConfig
	}{
		{name: "empty_name", config: llmguard.RegexDetectorConfig{Name: "", Entity: "X", Pattern: "a", Confidence: 0.5}},
		{name: "empty_entity", config: llmguard.RegexDetectorConfig{Name: "x", Entity: "", Pattern: "a", Confidence: 0.5}},
		{name: "empty_pattern", config: llmguard.RegexDetectorConfig{Name: "x", Entity: "X", Pattern: "", Confidence: 0.5}},
		{name: "high_confidence", config: llmguard.RegexDetectorConfig{Name: "x", Entity: "X", Pattern: "a", Confidence: 2}},
		{name: "nan_confidence", config: llmguard.RegexDetectorConfig{Name: "x", Entity: "X", Pattern: "a", Confidence: math.NaN()}},
		{name: "positive_inf_confidence", config: llmguard.RegexDetectorConfig{Name: "x", Entity: "X", Pattern: "a", Confidence: math.Inf(1)}},
		{name: "negative_inf_confidence", config: llmguard.RegexDetectorConfig{Name: "x", Entity: "X", Pattern: "a", Confidence: math.Inf(-1)}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := llmguard.NewCustomRegexpDetector(tc.config)
			require.Error(t, err)
			assert.ErrorIs(t, err, llmguard.ErrInvalidConfig)
		})
	}
}

func TestRegex_WhenDirectDetect_ExpectEmptyDetectorField(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-\d{6}`,
		Confidence: 0.9,
	})
	require.NoError(t, err)

	findings, err := detector.Detect(context.Background(), "EMP-123456")
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Empty(t, findings[0].Detector)
}

func TestRegex_WhenGuardDetect_ExpectRegisteredDetectorName(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-\d{6}`,
		Confidence: 0.9,
	})
	require.NoError(t, err)

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	findings, err := guard.Detect(context.Background(), "EMP-123456")
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "employee_id", findings[0].Detector)
}

func TestRegex_WhenUnicodeMatches_ExpectExactSpans(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-\d{6}`,
		Confidence: 0.9,
	})
	require.NoError(t, err)

	text := "значения EMP-123456 и EMP-654321 в тексте"
	findings, err := detector.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, 10, findings[0].End-findings[0].Start)
	assert.Equal(t, 10, findings[1].End-findings[1].Start)
}

func TestRegex_WhenZeroWidthPattern_ExpectNoFindings(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "zero",
		Entity:     llmguard.EntityType("ZERO"),
		Pattern:    `(?m)^`,
		Confidence: 0.5,
	})
	require.NoError(t, err)

	findings, err := detector.Detect(context.Background(), "line one\nline two")
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestRegex_WhenPatternInsideToken_ExpectCallerBoundary(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "digits",
		Entity:     llmguard.EntityType("DIGITS"),
		Pattern:    `\d{3}`,
		Confidence: 0.5,
	})
	require.NoError(t, err)

	text := "token123more"
	findings, err := detector.Detect(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, 3, findings[0].End-findings[0].Start)
}

func TestRegex_WhenConcurrentDetect_ExpectNoDataRace(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-\d+`,
		Confidence: 0.8,
	})
	require.NoError(t, err)

	text := "EMP-42"

	const workers = 16
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := detector.Detect(context.Background(), text)
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestRegex_WhenContextPreCanceled_ExpectCanceled(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-\d+`,
		Confidence: 0.8,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = detector.Detect(ctx, "EMP-42")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCustom_WhenRegexpRoundTrip_ExpectMaskRestore(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "employee_id",
		Entity:     llmguard.EntityType("EMPLOYEE_ID"),
		Pattern:    `EMP-[0-9]{6}`,
		Confidence: 0.9,
	})
	require.NoError(t, err)

	guard, err := llmguard.New(llmguard.WithDetector(detector))
	require.NoError(t, err)

	text := "worker EMP-123456 assigned"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, result.Findings)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

type customGoDetector struct {
	name string
}

func (d customGoDetector) Name() string { return d.name }

func (d customGoDetector) Detect(_ context.Context, text string) ([]llmguard.Finding, error) {
	idx := strings.Index(text, "CUST-1")
	if idx < 0 {
		return nil, nil
	}
	return []llmguard.Finding{{
		Entity:     llmguard.EntityType("CUSTOM_GO"),
		Start:      idx,
		End:        idx + len("CUST-1"),
		Confidence: 0.75,
		Detector:   d.name,
	}}, nil
}

func TestCustom_WhenGoDetectorRoundTrip_ExpectMaskRestore(t *testing.T) {
	t.Parallel()

	guard, err := llmguard.New(llmguard.WithDetector(customGoDetector{name: "custom_go"}))
	require.NoError(t, err)

	text := "value CUST-1 here"
	result, err := guard.Mask(context.Background(), text)
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)

	restored, err := guard.Restore(context.Background(), result.Text, result.Tokens)
	require.NoError(t, err)
	assert.Equal(t, text, restored)
}

func TestCustom_WhenOverlapWithBuiltIn_ExpectBuiltInWins(t *testing.T) {
	t.Parallel()

	detector, err := llmguard.NewCustomRegexpDetector(llmguard.RegexDetectorConfig{
		Name:       "wide_card",
		Entity:     llmguard.EntityType("CUSTOM_CARD"),
		Pattern:    `\d{16}`,
		Confidence: 1,
	})
	require.NoError(t, err)

	guard, err := llmguard.New(
		llmguard.WithDetector(detector),
		llmguard.WithDetector(llmguard.NewBankCardDetector()),
	)
	require.NoError(t, err)

	text := "card 4111111111111111"
	findings, err := guard.Detect(context.Background(), text)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	resolved, err := llmguard.Resolve(text, findings)
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntityBankCard, resolved[0].Entity)
}

func TestCustom_WhenUnsafeGoFinding_ExpectInvalidFindingError(t *testing.T) {
	t.Parallel()

	unsafe := &unsafeFindingDetector{name: "unsafe"}
	guard, err := llmguard.New(llmguard.WithDetector(unsafe))
	require.NoError(t, err)

	_, err = guard.Detect(context.Background(), "text")
	require.Error(t, err)
	assert.ErrorIs(t, err, llmguard.ErrInvalidFinding)
	assert.NotContains(t, err.Error(), "text")

	inv, ok := muerrors.As[*llmguard.InvalidFindingError](err)
	require.True(t, ok)
	assert.Equal(t, "span", inv.Field)
}

type unsafeFindingDetector struct {
	name string
}

func (d *unsafeFindingDetector) Name() string { return d.name }

func (d *unsafeFindingDetector) Detect(_ context.Context, text string) ([]llmguard.Finding, error) {
	return []llmguard.Finding{{
		Entity:     llmguard.EntityType("CUSTOM"),
		Start:      0,
		End:        len(text) + 1,
		Confidence: 0.5,
		Detector:   d.name,
	}}, nil
}

func TestResolve_WhenPassportOverlapsCustom_ExpectPassportWins(t *testing.T) {
	t.Parallel()

	payload := "паспорт 45 08 123456"
	passport := llmguard.Finding{
		Entity: llmguard.EntityPassport, Start: 8, End: 20, Confidence: 0.86, Detector: "passport",
	}
	custom := llmguard.Finding{
		Entity: llmguard.EntityType("CUSTOM_NUM"), Start: 8, End: 20, Confidence: 1, Detector: "custom",
	}
	resolved, err := llmguard.Resolve(payload, []llmguard.Finding{custom, passport})
	require.NoError(t, err)
	require.Len(t, resolved, 1)
	assert.Equal(t, llmguard.EntityPassport, resolved[0].Entity)
}
