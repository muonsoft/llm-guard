package evaluation_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateExposure_WhenSyntheticSuite_ExpectMetrics(t *testing.T) {
	t.Parallel()

	suite, err := evaluation.LoadSuite(filepath.Join(repoRoot(t), "testdata", "evaluation", "suites", "exposure.jsonl"))
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateExposure(context.Background(), guard, suite)
	require.NoError(t, err)
	assert.Equal(t, evaluation.StatusDiagnostic, report.Status)
	assert.Greater(t, report.Summary.SensitiveBytes, 0)
	require.NotEmpty(t, report.Ignored)
}

func TestFormatExposure_WhenMarkerInput_ExpectNoLeak(t *testing.T) {
	t.Parallel()

	suite, err := evaluation.LoadSuite(filepath.Join(repoRoot(t), "testdata", "evaluation", "suites", "exposure.jsonl"))
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateExposure(context.Background(), guard, suite)
	require.NoError(t, err)

	markdown := evaluation.FormatExposureMarkdown(report)
	jsonOut, err := evaluation.FormatExposureJSON(report)
	require.NoError(t, err)

	marker := evaluation.FormatSafeMarker
	assert.NotContains(t, markdown, marker)
	assert.NotContains(t, string(jsonOut), marker)
	assert.NotContains(t, markdown, "a@b.co")
	assert.NotContains(t, string(jsonOut), "a@b.co")
}

func TestFormatContract_WhenMarkerInput_ExpectNoLeak(t *testing.T) {
	t.Parallel()

	suite, err := evaluation.LoadSuite(filepath.Join(repoRoot(t), "testdata", "evaluation", "suites", "exposure.jsonl"))
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateContract(context.Background(), guard, suite)
	require.NoError(t, err)

	markdown := evaluation.FormatContractMarkdown(report)
	jsonOut, err := evaluation.FormatContractJSON(report)
	require.NoError(t, err)

	marker := evaluation.FormatSafeMarker
	assert.NotContains(t, markdown, marker)
	assert.NotContains(t, string(jsonOut), marker)
	assert.False(t, strings.Contains(markdown, "a@b.co"))
	assert.NotContains(t, string(jsonOut), "a@b.co")
}
