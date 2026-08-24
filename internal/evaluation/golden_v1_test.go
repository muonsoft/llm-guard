package evaluation_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluation_WhenV1GoldenMarkdown_ExpectStableFieldNames(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cases, err := evaluation.LoadCases(path)
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.Evaluate(context.Background(), guard, cases)
	require.NoError(t, err)
	require.False(t, report.HasRegression())

	markdown := evaluation.FormatMarkdown(report)
	for _, field := range []string{
		"| PERSON |", "| EMAIL |", "| CONNECTION_STRING |",
		"Aggregate TP=", "Coverage complete: true",
	} {
		assert.Contains(t, markdown, field)
	}
}

func TestEvaluation_WhenV1GoldenJSON_ExpectStableFieldNames(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cases, err := evaluation.LoadCases(path)
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.Evaluate(context.Background(), guard, cases)
	require.NoError(t, err)

	jsonReport, err := evaluation.FormatJSON(report)
	require.NoError(t, err)
	body := string(jsonReport)
	for _, field := range []string{
		`"entity"`, `"tp"`, `"fp"`, `"fn"`,
		`"negative_cases"`, `"false_positive_cases"`,
		`"precision"`, `"recall"`, `"f1"`, `"fpr"`, `"fnr"`,
		`"coverage"`, `"complete"`, `"has_positive"`, `"has_negative"`,
	} {
		assert.Contains(t, body, field)
	}
}

func TestEvaluation_WhenV1EntityOrder_ExpectMVPSequence(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cases, err := evaluation.LoadCases(path)
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.Evaluate(context.Background(), guard, cases)
	require.NoError(t, err)
	require.Len(t, report.Entities, evaluation.MVPEntityCount())
	assert.Equal(t, "PERSON", string(report.Entities[0].Entity))
	assert.Equal(t, "CONNECTION_STRING", string(report.Entities[len(report.Entities)-1].Entity))
}

func TestEvaluationCLI_WhenV1RegressionGate_ExpectExitZero(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	corpus := filepath.Join(root, "testdata", "evaluation", "cases.jsonl")
	cmd := exec.Command("go", "run", "./cmd/llmguard-eval",
		"-corpus", corpus,
		"-format", "markdown",
		"-fail-on-regression",
	)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestEvaluationCLI_WhenV1JSONRegressionGate_ExpectExitZero(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	corpus := filepath.Join(root, "testdata", "evaluation", "cases.jsonl")
	cmd := exec.Command("go", "run", "./cmd/llmguard-eval",
		"-corpus", corpus,
		"-format", "json",
		"-fail-on-regression",
	)
	cmd.Dir = root
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(out), "{\n"))
}
