package evaluation_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluation_WhenSchemaInvalid_ExpectRejection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(`{"schema_version":99,"id":"x","category":"positive","input":"a","expected":[]}`+"\n"), 0o600))

	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema_version")
}

func TestEvaluation_WhenUnknownField_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"negative","input":"a","expected":[],"extra":1}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestEvaluation_WhenTrailingJSON_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"negative","input":"a","expected":[]}{}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "trailing JSON") || strings.Contains(err.Error(), "after top-level value"))
}

func TestEvaluation_WhenUnknownEntity_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"positive","input":"abc","expected":[{"entity":"UNKNOWN","start":0,"end":1}]}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown entity")
}

func TestEvaluation_WhenDuplicateSpan_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"positive","input":"abc","expected":[{"entity":"EMAIL","start":0,"end":1},{"entity":"EMAIL","start":0,"end":1}]}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate span")
}

func TestEvaluation_WhenPositiveWithoutExpected_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"positive","input":"abc","expected":[]}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "positive case requires expected spans")
}

func TestEvaluation_WhenNegativeWithExpected_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"negative","input":"abc","expected":[{"entity":"EMAIL","start":0,"end":1}]}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "negative case must not include expected spans")
}

func TestEvaluation_WhenMixedSingleEntity_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"x","category":"mixed","input":"abc","expected":[{"entity":"EMAIL","start":0,"end":1}]}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixed case requires at least two distinct expected entities")
}

func TestEvaluation_WhenMissingEntityCoverage_ExpectRejection(t *testing.T) {
	t.Parallel()

	path := writeCorpusLine(t, `{"schema_version":1,"id":"email-positive","category":"positive","input":"a@b.co","expected":[{"entity":"EMAIL","start":0,"end":6}]}`)
	_, err := evaluation.LoadCases(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "coverage incomplete")
}

func TestEvaluation_WhenSpanMismatch_ExpectFPFN(t *testing.T) {
	t.Parallel()

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	cases := []evaluation.Case{{
		SchemaVersion: evaluation.SchemaVersion,
		ID:            "email-positive",
		Category:      "positive",
		Input:         "mail a@b.co",
		Expected:      []evaluation.ExpectedSpan{{Entity: "EMAIL", Start: 0, End: 4}},
	}}
	report, err := evaluation.Evaluate(context.Background(), guard, cases)
	require.NoError(t, err)
	assert.True(t, report.HasRegression())
}

func TestEvaluation_WhenCommittedCorpus_ExpectDeterministicReport(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cases, err := evaluation.LoadCases(path)
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.Evaluate(context.Background(), guard, cases)
	require.NoError(t, err)
	assert.False(t, report.HasRegression())
	assert.True(t, report.Coverage.Complete)

	markdown := evaluation.FormatMarkdown(report)
	assert.Contains(t, markdown, "| PERSON |")
	assert.Contains(t, markdown, "| CONNECTION_STRING |")
	assert.Contains(t, markdown, "Coverage complete: true")

	jsonReport, err := evaluation.FormatJSON(report)
	require.NoError(t, err)
	assert.Contains(t, string(jsonReport), `"entity": "PERSON"`)
	assert.Contains(t, string(jsonReport), `"coverage"`)
	assert.Contains(t, string(jsonReport), `"complete": true`)
	assert.Contains(t, string(jsonReport), `"has_positive"`)
	assert.Contains(t, string(jsonReport), `"has_negative"`)
}

func TestEvaluation_WhenJSONFormattedIncomplete_ExpectCoverageEvidence(t *testing.T) {
	t.Parallel()

	report := evaluation.Report{
		Cases: 1,
		Coverage: evaluation.CoverageReport{
			Complete: false,
			Entities: []evaluation.EntityCoverage{{
				Entity:      "PERSON",
				HasPositive: true,
				HasNegative: false,
			}},
		},
	}
	jsonReport, err := evaluation.FormatJSON(report)
	require.NoError(t, err)
	assert.Contains(t, string(jsonReport), `"complete": false`)
	assert.Contains(t, string(jsonReport), `"has_positive": true`)
	assert.Contains(t, string(jsonReport), `"has_negative": false`)
}

func TestEvaluation_WhenCommittedCorpus_ExpectEntityCoverageForAllMVP(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repoRoot(t), "testdata", "evaluation", "cases.jsonl")
	cases, err := evaluation.LoadCases(path)
	require.NoError(t, err)

	coverage := evaluation.BuildCoverage(cases)
	require.True(t, coverage.Complete)
	require.Len(t, coverage.Entities, evaluation.MVPEntityCount())
	for _, row := range coverage.Entities {
		assert.True(t, row.HasPositive, "entity=%s", row.Entity)
		assert.True(t, row.HasNegative, "entity=%s", row.Entity)
	}
}

func TestEvaluation_WhenZeroDenominator_ExpectZeroRates(t *testing.T) {
	t.Parallel()

	report := evaluation.Report{
		Entities: []evaluation.EntityMetrics{{
			Entity: "PERSON",
		}},
		Coverage: evaluation.CoverageReport{Complete: true},
	}
	markdown := evaluation.FormatMarkdown(report)
	assert.Contains(t, markdown, "| PERSON | 0 | 0 | 0 |")
}

func writeCorpusLine(t *testing.T, line string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(strings.TrimSpace(line)+"\n"), 0o600))
	return path
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
