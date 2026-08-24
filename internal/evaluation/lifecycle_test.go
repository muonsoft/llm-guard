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

func lifecycleGuard(t *testing.T) *llmguard.Guard {
	t.Helper()
	guard, err := evaluation.NewDeterministicGuard(nil)
	require.NoError(t, err)
	return guard
}

func TestEvaluateLifecycle_WhenSmokeSuiteIdentity_ExpectPass(t *testing.T) {
	t.Parallel()
	suite, err := evaluation.LoadSuite(filepath.Join(repoRoot(t), "testdata", "evaluation", "generated", "smoke.jsonl"))
	require.NoError(t, err)
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), suite)
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression(), "%+v", report.Diagnostics)
}

func TestEvaluateLifecycle_WhenMutationRecipe_ExpectNoRegression(t *testing.T) {
	t.Parallel()
	suite := evaluation.Suite{
		SuiteID:        "s",
		MappingVersion: "v1",
		SourceIDs:      []string{"generated"},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:  evaluation.SuiteSchemaVersion,
			SuiteID:        "s",
			SourceID:       "generated",
			SourceRecordID: "mut",
			MappingVersion: "v1",
			Input:          "Contact a@b.co",
			InputSHA256:    "ab81dd8ef48b30ede16d7e2655e7b0019937da4d05a4e697dba6c97e5e41196f",
			Annotations: []evaluation.SuiteAnnotation{{
				SourceLabel: "EMAIL", MappedEntity: "EMAIL", Start: 8, End: 14, Disposition: evaluation.DispositionSupported,
			}},
			Lifecycle: &evaluation.SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "mutate_placeholder"},
		}},
	}
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), suite)
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
}

func TestLifecycleReport_WhenMarkerInInput_ExpectNotInMarkdown(t *testing.T) {
	t.Parallel()
	report := evaluation.LifecycleReport{
		Profile: "lifecycle",
		SuiteID: "s",
		Status:  evaluation.StatusPass,
		Diagnostics: []evaluation.LifecycleFailureDiagnostic{{
			SourceRecordID: "r1",
			ExpectedAction: "mask",
			Outcome:        evaluation.LifecycleOutcomeRestoreMiss,
			Detail:         "restore mismatch",
			InputSHA256:    "abc",
		}},
	}
	out := evaluation.FormatLifecycleMarkdown(report)
	assert.NotContains(t, out, evaluation.FormatSafeMarker)
}
