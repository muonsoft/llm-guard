package evaluation_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/muonsoft/llm-guard"
	"github.com/muonsoft/llm-guard/internal/evaluation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateContract_WhenSyntheticSuite_ExpectPass(t *testing.T) {
	t.Parallel()

	suite, err := evaluation.LoadSuite(filepath.Join(repoRoot(t), "testdata", "evaluation", "suites", "contract.jsonl"))
	require.NoError(t, err)

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateContract(context.Background(), guard, suite)
	require.NoError(t, err)
	assert.False(t, report.HasContractRegression())
	assert.Equal(t, evaluation.StatusPass, report.Status)
	require.Len(t, report.Entities, 1)
	assert.Equal(t, "EMAIL", string(report.Entities[0].Entity))
}

func TestEvaluateContract_WhenPartialOverlap_ExpectNotTP(t *testing.T) {
	t.Parallel()

	suite := evaluation.Suite{
		SuiteID:        "partial",
		MappingVersion: "v1",
		SourceIDs:      []string{"s"},
		Scope:          []llmguard.EntityType{llmguard.EntityEmail},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:  evaluation.SuiteSchemaVersion,
			SuiteID:        "partial",
			SourceID:       "s",
			SourceRecordID: "r1",
			MappingVersion: "v1",
			Input:          "Contact a@b.co",
			InputSHA256:    "ab81dd8ef48b30ede16d7e2655e7b0019937da4d05a4e697dba6c97e5e41196f",
			Annotations: []evaluation.SuiteAnnotation{{
				SourceLabel:  "EMAIL",
				MappedEntity: "EMAIL",
				Start:        8,
				End:          12,
				Disposition:  evaluation.DispositionSupported,
			}},
		}},
	}

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateContract(context.Background(), guard, suite)
	require.NoError(t, err)
	assert.True(t, report.HasContractRegression())
	assert.Greater(t, report.Diagnostics.OverlappingPairs, 0)
	require.NotEmpty(t, report.Diagnostics.Failures)
	assert.NotContains(t, report.Diagnostics.Failures[0].SourceRecordID, "a@b.co")

	markdown := evaluation.FormatContractMarkdown(report)
	jsonOut, err := evaluation.FormatContractJSON(report)
	require.NoError(t, err)
	assert.Contains(t, markdown, "Failure diagnostics")
	assert.Contains(t, string(jsonOut), `"failures"`)
	assert.Contains(t, string(jsonOut), `"input_sha256"`)
	assert.NotContains(t, markdown, "a@b.co")
}

func TestEvaluateContract_WhenSubsetScope_ExpectOnlyDeclaredEntities(t *testing.T) {
	t.Parallel()

	suite := evaluation.Suite{
		SuiteID:        "subset",
		MappingVersion: "v1",
		SourceIDs:      []string{"s"},
		Scope:          []llmguard.EntityType{llmguard.EntityEmail},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:    evaluation.SuiteSchemaVersion,
			SuiteID:          "subset",
			SourceID:         "s",
			SourceRecordID:   "r1",
			MappingVersion:   "v1",
			Input:            "Contact a@b.co",
			InputSHA256:      "ab81dd8ef48b30ede16d7e2655e7b0019937da4d05a4e697dba6c97e5e41196f",
			DeclaredEntities: []string{"EMAIL"},
			Annotations: []evaluation.SuiteAnnotation{
				{SourceLabel: "EMAIL", MappedEntity: "EMAIL", Start: 8, End: 14, Disposition: evaluation.DispositionSupported},
				{SourceLabel: "PHONE", MappedEntity: "PHONE", Start: 0, End: 7, Disposition: evaluation.DispositionSupported},
			},
		}},
	}

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateContract(context.Background(), guard, suite)
	require.NoError(t, err)
	require.Len(t, report.Entities, 1)
	assert.Equal(t, "EMAIL", string(report.Entities[0].Entity))
}

func TestEvaluateContract_WhenZeroDenominator_ExpectZeroRates(t *testing.T) {
	t.Parallel()

	suite := evaluation.Suite{
		SuiteID:        "empty-scope",
		MappingVersion: "v1",
		SourceIDs:      []string{"s"},
		Scope:          []llmguard.EntityType{llmguard.EntityEmail},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:  evaluation.SuiteSchemaVersion,
			SuiteID:        "empty-scope",
			SourceID:       "s",
			SourceRecordID: "r1",
			MappingVersion: "v1",
			Input:          "no entities",
			InputSHA256:    "aa191de6a32acc53ad63c2b884143130bbb9b8316669f88883177e2d1f96f264",
			Annotations:    nil,
		}},
	}

	guard, err := evaluation.NewMVPGuard()
	require.NoError(t, err)

	report, err := evaluation.EvaluateContract(context.Background(), guard, suite)
	require.NoError(t, err)
	require.Len(t, report.Entities, 1)
	row := report.Entities[0]
	assert.Equal(t, 0.0, row.Precision)
	assert.Equal(t, 0.0, row.Recall)
	assert.Equal(t, 0.0, row.F1)
	assert.Equal(t, 0.0, row.FPR)
	assert.Equal(t, 0.0, row.FNR)
}
