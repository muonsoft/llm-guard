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

func cyrillicFIOEmailSuite(recipe string) evaluation.Suite {
	return evaluation.Suite{
		SuiteID:        "s",
		MappingVersion: "v1",
		SourceIDs:      []string{"generated"},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:  evaluation.SuiteSchemaVersion,
			SuiteID:        "s",
			SourceID:       "generated",
			SourceRecordID: "cyrillic-fio-email",
			MappingVersion: "v1",
			Input:          "Иван Петров a@b.co",
			InputSHA256:    "placeholder-sha256",
			Annotations: []evaluation.SuiteAnnotation{
				{SourceLabel: "PERSON", MappedEntity: "PERSON", Start: 0, End: 21, Disposition: evaluation.DispositionSupported},
				{SourceLabel: "EMAIL", MappedEntity: "EMAIL", Start: 22, End: 28, Disposition: evaluation.DispositionSupported},
			},
			Lifecycle: &evaluation.SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: recipe},
		}},
	}
}

func literalEmailSuite(recipe string) evaluation.Suite {
	literal := "{{LLMG_ffffffffffffffffffffffffffffffff_9999}}"
	input := literal + " a@b.co"
	return evaluation.Suite{
		SuiteID:        "s",
		MappingVersion: "v1",
		SourceIDs:      []string{"generated"},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:  evaluation.SuiteSchemaVersion,
			SuiteID:        "s",
			SourceID:       "generated",
			SourceRecordID: "literal-email",
			MappingVersion: "v1",
			Input:          input,
			InputSHA256:    "placeholder-sha256",
			Annotations: []evaluation.SuiteAnnotation{
				{SourceLabel: "EMAIL", MappedEntity: "EMAIL", Start: len(literal) + 1, End: len(input), Disposition: evaluation.DispositionSupported},
			},
			Lifecycle: &evaluation.SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: recipe},
		}},
	}
}

func TestEvaluateLifecycle_WhenPreservedLiteralEmailIdentity_ExpectNoRegression(t *testing.T) {
	t.Parallel()
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), literalEmailSuite("identity"))
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
	assert.Empty(t, report.Diagnostics)
}

func TestEvaluateLifecycle_WhenPreservedLiteralEmailMutate_ExpectNoRegression(t *testing.T) {
	t.Parallel()
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), literalEmailSuite("mutate_placeholder"))
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
	assert.GreaterOrEqual(t, report.MutationMiss, 1)
	assert.Empty(t, report.Diagnostics)
}

func TestEvaluateLifecycle_WhenPreservedLiteralEmailDelete_ExpectNoRegression(t *testing.T) {
	t.Parallel()
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), literalEmailSuite("delete_placeholder"))
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
	assert.GreaterOrEqual(t, report.MutationMiss, 1)
	assert.Empty(t, report.Diagnostics)
}

func TestEvaluateLifecycle_WhenCyrillicFIOEmailMutate_ExpectNoRegression(t *testing.T) {
	t.Parallel()
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), cyrillicFIOEmailSuite("mutate_placeholder"))
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
	assert.GreaterOrEqual(t, report.MutationMiss, 1)
	assert.Empty(t, report.Diagnostics)
}

func TestEvaluateLifecycle_WhenCyrillicFIOEmailDelete_ExpectNoRegression(t *testing.T) {
	t.Parallel()
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), cyrillicFIOEmailSuite("delete_placeholder"))
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
	assert.GreaterOrEqual(t, report.MutationMiss, 1)
	assert.Empty(t, report.Diagnostics)
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
	assert.GreaterOrEqual(t, report.MutationMiss, 1)
	assert.Empty(t, report.Diagnostics)
}

func TestEvaluateLifecycle_WhenCollisionRecipe_ExpectFakeTokenPreserved(t *testing.T) {
	t.Parallel()
	suite := evaluation.Suite{
		SuiteID:        "s",
		MappingVersion: "v1",
		SourceIDs:      []string{"generated"},
		Records: []evaluation.SuiteRecord{{
			SchemaVersion:  evaluation.SuiteSchemaVersion,
			SuiteID:        "s",
			SourceID:       "generated",
			SourceRecordID: "col",
			MappingVersion: "v1",
			Input:          "Contact a@b.co",
			InputSHA256:    "ab81dd8ef48b30ede16d7e2655e7b0019937da4d05a4e697dba6c97e5e41196f",
			Annotations: []evaluation.SuiteAnnotation{{
				SourceLabel: "EMAIL", MappedEntity: "EMAIL", Start: 8, End: 14, Disposition: evaluation.DispositionSupported,
			}},
			Lifecycle: &evaluation.SuiteLifecycle{ExpectedAction: "mask", ResponseRecipe: "collision"},
		}},
	}
	report, err := evaluation.EvaluateLifecycle(context.Background(), lifecycleGuard(t), suite)
	require.NoError(t, err)
	assert.False(t, report.HasLifecycleRegression())
	assert.GreaterOrEqual(t, report.MutationMiss, 1)
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
